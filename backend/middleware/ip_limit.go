package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"server/config"
	"server/utils"
)

// IPLimiter 限制同时访问的唯一 IP 数量。
type IPLimiter struct {
	mu     sync.RWMutex
	active map[string]int // IP -> 当前活跃请求数
	maxIPs int
}

// NewIPLimiter 创建一个新的 IP 并发限制器。
func NewIPLimiter() *IPLimiter {
	maxIPs := config.AppConfig.Server.MaxConcurrentIPs
	if maxIPs <= 0 {
		maxIPs = 10
	}
	return &IPLimiter{
		active: make(map[string]int),
		maxIPs: maxIPs,
	}
}

// Limit 返回限制同时访问 IP 数量的 Gin 中间件。
func (l *IPLimiter) Limit() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()

		l.mu.Lock()
		count, exists := l.active[ip]
		if !exists && len(l.active) >= l.maxIPs {
			l.mu.Unlock()
			ctx.JSON(http.StatusOK, utils.Error(1, "访问人数过多，请稍后再试"))
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
