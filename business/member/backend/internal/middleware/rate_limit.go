package middleware

import (
	"sync"
	"time"

	"member/pkg/response"

	"github.com/gin-gonic/gin"
)

// IPRateLimiter 基于 IP 的固定窗口限速器，带周期清理防止内存泄漏。
type IPRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateEntry
	limit   int
	window  time.Duration
	cleanAt time.Time
}

type rateEntry struct {
	count     int
	windowEnd time.Time
}

func NewIPRateLimiter(limit int, window time.Duration) *IPRateLimiter {
	return &IPRateLimiter{
		entries: make(map[string]*rateEntry),
		limit:   limit,
		window:  window,
		cleanAt: time.Now().Add(time.Minute),
	}
}

// Allow 判断该 IP 是否允许继续请求。
func (l *IPRateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	l.cleanup(now)

	e, ok := l.entries[ip]
	if !ok || now.After(e.windowEnd) {
		l.entries[ip] = &rateEntry{count: 1, windowEnd: now.Add(l.window)}
		return true
	}
	e.count++
	return e.count <= l.limit
}

// cleanup 每分钟清理一次已过期的记录，避免 map 无限增长。
func (l *IPRateLimiter) cleanup(now time.Time) {
	if now.Before(l.cleanAt) {
		return
	}
	l.cleanAt = now.Add(time.Minute)
	for ip, e := range l.entries {
		if now.After(e.windowEnd) {
			delete(l.entries, ip)
		}
	}
}

// RateLimitByIP 生成按 IP 限速的中间件。
func RateLimitByIP(limiter *IPRateLimiter, msg string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow(c.ClientIP()) {
			response.Error(c, 429, msg)
			c.Abort()
			return
		}
		c.Next()
	}
}
