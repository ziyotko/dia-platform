package middleware

import (
	"member/pkg/response"
	"sync"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

type ipLimiter struct {
	mu            sync.RWMutex
	ipCounts      map[string]*int32
	maxConcurrent int
}

var limiter *ipLimiter

func InitIPLimiter(max int) {
	limiter = &ipLimiter{
		ipCounts:      make(map[string]*int32),
		maxConcurrent: max,
	}
}

func IPLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter == nil {
			c.Next()
			return
		}

		ip := c.ClientIP()
		limiter.mu.RLock()
		count, exists := limiter.ipCounts[ip]
		limiter.mu.RUnlock()

		if !exists {
			limiter.mu.Lock()
			count = new(int32)
			limiter.ipCounts[ip] = count
			limiter.mu.Unlock()
		}

		current := atomic.AddInt32(count, 1)
		defer atomic.AddInt32(count, -1)

		if int(current) > limiter.maxConcurrent {
			response.Error(c, 429, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}
