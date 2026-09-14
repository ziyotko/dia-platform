package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"application/pkg/redis"
	"application/pkg/response"

	"github.com/gin-gonic/gin"
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

// RateLimitMiddleware 基于 Redis 的固定窗口限流，按客户端真实 IP 计数（与 portal 的 middleware/rate_limit.go 保持一致）。
// 用于验证码等公开接口，防止被脚本刷量/防暴力破解。
//
// 窗口 key 以 IP + 当前时间窗口段生成，便于多实例共享计数；Redis 不可用时退化为进程内限流（fail-closed）。
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	if limit <= 0 {
		limit = 60
	}
	if window <= 0 {
		window = time.Minute
	}
	return func(c *gin.Context) {
		ip := c.ClientIP()
		windowSec := int64(window.Seconds())
		key := fmt.Sprintf("ratelimit:%s:%d", ip, time.Now().Unix()/windowSec)

		count, err := redis.AntiReplayClient.Incr(redis.Ctx, key).Result()
		if err != nil {
			// Redis 不可用：退化为进程内限流兜底（fail-closed），避免限流组件故障时被无限刷量
			if !allowLocalRateLimit("ratelimit:"+ip, limit, window) {
				response.FailWithCode(c, http.StatusTooManyRequests, "请求过于频繁，请稍后再试")
				c.Abort()
				return
			}
			c.Next()
			return
		}
		if count == 1 {
			redis.AntiReplayClient.Expire(redis.Ctx, key, window)
		} else if redis.AntiReplayClient.TTL(redis.Ctx, key).Val() < 0 {
			// 兜底：进程中途异常可能留下无 TTL 的 key，避免其永久驻留
			redis.AntiReplayClient.Expire(redis.Ctx, key, window)
		}
		if count > int64(limit) {
			response.FailWithCode(c, http.StatusTooManyRequests, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}
		c.Next()
	}
}
