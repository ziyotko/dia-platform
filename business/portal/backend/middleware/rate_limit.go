package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"server/utils"
)

// localRateLimitEntry 进程内固定窗口限流条目
type localRateLimitEntry struct {
	windowStart int64
	count       int64
}

var (
	localRateLimitMu    sync.Mutex
	localRateLimitStore = make(map[string]*localRateLimitEntry)
)

// allowLocalRateLimit 进程内固定窗口限流兜底：Redis 不可用时防止限流完全失效（fail-closed）。
// 仅在单实例内生效，作为降级方案，达不到分布式精确限流的效果但可挡住脚本刷量。
func allowLocalRateLimit(key string, limit int, window time.Duration) bool {
	windowSec := int64(window.Seconds())
	if windowSec <= 0 {
		windowSec = 60
	}
	now := time.Now().Unix()
	windowStart := now - now%windowSec

	localRateLimitMu.Lock()
	defer localRateLimitMu.Unlock()

	entry, ok := localRateLimitStore[key]
	if !ok || entry.windowStart != windowStart {
		entry = &localRateLimitEntry{windowStart: windowStart}
		localRateLimitStore[key] = entry
	}
	entry.count++

	// 机会式清理：避免 map 无限增长（仅在条目较多时触发）
	if len(localRateLimitStore) > 10000 {
		for k, e := range localRateLimitStore {
			if e.windowStart != windowStart {
				delete(localRateLimitStore, k)
			}
		}
	}
	return entry.count <= int64(limit)
}

// RateLimitMiddleware 基于 Redis 的固定窗口限流，按真实客户端 IP 计数。
// 用于公开写接口（如站点分析），防止被脚本刷量/灌脏数据。
//
// 窗口 key 以 IP + 当前时间窗口段生成，便于分布式场景下多实例共享计数。
// 返回 limit 个请求/窗口 之外的请求被拒绝。
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	if limit <= 0 {
		limit = 60
	}
	if window <= 0 {
		window = time.Minute
	}
	return func(ctx *gin.Context) {
		ip := utils.RealIP(ctx)
		windowSec := int64(window.Seconds())
		now := time.Now().Unix()
		key := fmt.Sprintf("ratelimit:%s:%d", ip, now/windowSec)

		count, err := utils.Redis1.Incr(utils.Ctx, key).Result()
		if err != nil {
			// Redis 不可用：退化为进程内限流兜底（fail-closed），避免限流组件故障时被无限刷量
			if !allowLocalRateLimit("ratelimit:"+ip, limit, window) {
				ctx.JSON(http.StatusOK, utils.Error(1, "请求过于频繁，请稍后再试"))
				ctx.Abort()
				return
			}
			ctx.Next()
			return
		}
		if count == 1 {
			utils.Redis1.Expire(utils.Ctx, key, window)
		} else if utils.Redis1.TTL(utils.Ctx, key).Val() < 0 {
			// 兜底：进程中途异常可能留下无 TTL 的 key，避免其永久驻留
			utils.Redis1.Expire(utils.Ctx, key, window)
		}
		if count > int64(limit) {
			ctx.JSON(http.StatusOK, utils.Error(1, "请求过于频繁，请稍后再试"))
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
