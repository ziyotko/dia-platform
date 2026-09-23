package middleware

import (
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"application/pkg/response"

	"github.com/gin-gonic/gin"
)

var (
	ipCounts = make(map[string]*int32)
	ipMutex  sync.RWMutex
	maxPerIP int32 = 100
)

// defaultMaxConcurrentIPs 与 config.yaml 的默认值一致。
const defaultMaxConcurrentIPs = 100

// InitIPLimiter 初始化单 IP 并发上限。
// 配成 0 或负数时回退到默认值：否则 `current > 0` 恒成立，全站请求都会被 429。
func InitIPLimiter(maxConcurrent int) {
	if maxConcurrent <= 0 {
		log.Printf("[WARN] server.max_concurrent_ips = %d 无效，回退为默认值 %d（配成 0 会让所有请求都返回 429）", maxConcurrent, defaultMaxConcurrentIPs)
		maxConcurrent = defaultMaxConcurrentIPs
	}
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
		defer func() {
			// 计数归零就把条目从 map 里删掉：否则长时间运行后 ipCounts 会无限增长
			// （只删指针仍然相同且计数为 0 的条目，避免误删并发中的新计数）
			if atomic.AddInt32(counter, -1) == 0 {
				ipMutex.Lock()
				if cur, ok := ipCounts[ip]; ok && cur == counter && atomic.LoadInt32(cur) == 0 {
					delete(ipCounts, ip)
				}
				ipMutex.Unlock()
			}
		}()

		if current > maxPerIP {
			response.FailWithCode(c, http.StatusTooManyRequests, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}
		c.Next()
	}
}
