package middleware

import (
	"fmt"
	"sync"
	"time"

	"base/pkg/redis"
	"base/pkg/response"

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

// RateLimitMiddleware 基于 Redis 的固定窗口限流，按「接口 + 真实客户端 IP」计数
// （实现参考 portal / member / application 的 middleware/rate_limit.go）。
//
// 用于公开接口（登录、验证码、初始化超管），防止暴力破解、刷验证码与脚本刷量。
// 窗口 key 以路由模板 + IP + 当前时间窗口段生成，多实例共享计数；Redis 不可用时退化为进程内限流。
// 超过 limit 次/窗口的请求返回「请求过于频繁，请稍后再试」。
//
// 与 portal 当前实现的差异：portal 的 key 不含路由（`ratelimit:<ip>:<窗口段>`），
// 所以同一 IP 下各接口共用一个计数，验证码刷多了会把登录接口的额度一并吃掉；
// 这里加上路由模板，使各接口独立计算各自的配额。
//
// 注意：客户端 IP 由 gin 的 ClientIP() 解析，仅当请求来自 server.trusted_proxies
// 中配置的可信代理时才信任 X-Forwarded-For / X-Real-IP，避免伪造头绕过限流。
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	if limit <= 0 {
		limit = 60
	}
	if window <= 0 {
		window = time.Minute
	}
	windowSec := int64(window.Seconds())

	return func(c *gin.Context) {
		ip := c.ClientIP()
		// 路由模板（如 /base/api/v1/auth/login），未匹配时退化为实际路径
		scope := c.FullPath()
		if scope == "" {
			scope = c.Request.URL.Path
		}
		now := time.Now().Unix()
		key := fmt.Sprintf("ratelimit:%s:%s:%d", scope, ip, now/windowSec)

		count, err := redis.Client.Incr(redis.Ctx, key).Result()
		if err != nil {
			// Redis 不可用：退化为进程内限流兜底（fail-closed），避免限流组件故障时被无限刷量
			if !allowLocalRateLimit("local:"+key, limit, window) {
				response.FailWithCode(c, response.CodeTooManyRequests, "请求过于频繁，请稍后再试")
				c.Abort()
				return
			}
			c.Next()
			return
		}
		if count == 1 {
			_ = redis.Client.Expire(redis.Ctx, key, window).Err()
		} else if redis.Client.TTL(redis.Ctx, key).Val() < 0 {
			// 兜底：进程中途异常可能留下无 TTL 的 key，避免其永久驻留
			_ = redis.Client.Expire(redis.Ctx, key, window).Err()
		}
		if count > int64(limit) {
			response.FailWithCode(c, response.CodeTooManyRequests, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}
		c.Next()
	}
}
