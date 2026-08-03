package middleware

import (
	"net/http"
	"sync"
	"sync/atomic"

	"conference/pkg/response"

	"github.com/gin-gonic/gin"
)

var (
	ipCounts = make(map[string]*int32)
	ipMutex  sync.RWMutex
	maxPerIP int32 = 100
)

func InitIPLimiter(maxConcurrent int) {
	maxPerIP = int32(maxConcurrent)
}

func IPLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		ipMutex.RLock()
		counter, exists := ipCounts[ip]
		ipMutex.RUnlock()

		if !exists {
			ipMutex.Lock()
			if _, exists = ipCounts[ip]; !exists {
				var c int32
				ipCounts[ip] = &c
			}
			counter = ipCounts[ip]
			ipMutex.Unlock()
		}

		current := atomic.AddInt32(counter, 1)
		defer atomic.AddInt32(counter, -1)

		if current > maxPerIP {
			response.FailWithCode(c, http.StatusTooManyRequests, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}
		c.Next()
	}
}
