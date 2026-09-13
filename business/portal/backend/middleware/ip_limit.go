package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"server/config"
	"server/utils"
)

// IPLimiter 限制单个 IP 的并发请求数（防止单 IP 用大量连接占满资源）。
// 说明：不再按“全站唯一 IP 数”拒绝，旧算法会让攻击者用大量源 IP 占满上限、把合法用户全部挡在门外。
type IPLimiter struct {
	mu            sync.RWMutex
	active        map[string]int // IP -> 当前活跃请求数
	maxConcurrent int
}

// NewIPLimiter 创建一个新的 IP 并发限制器。
func NewIPLimiter() *IPLimiter {
	maxConcurrent := config.AppConfig.Server.MaxConcurrentIPs
	if maxConcurrent <= 0 {
		maxConcurrent = 10
	}
	return &IPLimiter{
		active:        make(map[string]int),
		maxConcurrent: maxConcurrent,
	}
}

// Limit 返回限制单 IP 并发请求数的 Gin 中间件。
func (l *IPLimiter) Limit() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ip := utils.RealIP(ctx)

		l.mu.Lock()
		count := l.active[ip]
		if count >= l.maxConcurrent {
			l.mu.Unlock()
			ctx.JSON(http.StatusOK, utils.Error(1, "访问过于频繁，请稍后再试"))
			ctx.Abort()
			return
		}
		l.active[ip] = count + 1
		l.mu.Unlock()

		defer func() {
			l.mu.Lock()
			l.active[ip]--
			if l.active[ip] <= 0 {
				delete(l.active, ip)
			}
			l.mu.Unlock()
		}()

		ctx.Next()
	}
}
