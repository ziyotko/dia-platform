package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"server/config"
)

// CorsMiddleware 仅放行 config.yaml 中 allowed_origins 白名单内的来源，
// 避免 Access-Control-Allow-Origin:* 带来的跨域风险。
func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && isAllowedOrigin(origin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-Request-Timestamp, X-Request-Nonce, X-Request-Signature, X-Body-Hash-Value")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func isAllowedOrigin(origin string) bool {
	for _, o := range config.AppConfig.Server.AllowedOrigins {
		if strings.TrimSpace(o) == origin {
			return true
		}
	}
	return false
}
