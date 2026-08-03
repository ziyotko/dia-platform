package middleware

import (
	"member/pkg/response"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

type ipLimiter struct {
	mu            sync.Mutex
	ipCounts      map[string]*int32
	lastSeen      map[string]time.Time
	maxConcurrent int
	stop          chan struct{}
}

var limiter *ipLimiter

func InitIPLimiter(max int) {
	limiter = &ipLimiter{
		ipCounts:      make(map[string]*int32),
		lastSeen:      make(map[string]time.Time),
		maxConcurrent: max,
		stop:          make(chan struct{}),
	}
	go limiter.cleanupLoop()
}

// cleanupLoop 定期清理长期无活动的 IP 记录，防止内存泄漏。
func (l *ipLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			l.cleanup(time.Now())
		case <-l.stop:
			return
		}
	}
}

func (l *ipLimiter) cleanup(now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for ip, ts := range l.lastSeen {
		if now.Sub(ts) > 10*time.Minute {
			delete(l.ipCounts, ip)
			delete(l.lastSeen, ip)
		}
	}
}

func IPLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter == nil {
			c.Next()
			return
		}

		ip := c.ClientIP()
		limiter.mu.Lock()
		count, exists := limiter.ipCounts[ip]
		if !exists {
			count = new(int32)
			limiter.ipCounts[ip] = count
		}
		limiter.lastSeen[ip] = time.Now()
		limiter.mu.Unlock()

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
