package middleware

import (
	"member/config"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS 仅放行 config.yaml 中 allowed_origins 白名单内的来源，
// 避免 Access-Control-Allow-Origin:* 与凭据混用带来的风险。
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && isAllowedOrigin(origin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin,Content-Type,Accept,Authorization,X-Request-Timestamp,X-Request-Nonce,X-Request-Signature")
			c.Header("Access-Control-Expose-Headers", "Content-Length,Content-Disposition")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Vary", "Origin")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func isAllowedOrigin(origin string) bool {
	for _, o := range config.Cfg.Server.AllowedOrigins {
		if strings.TrimSpace(o) == origin {
			return true
		}
	}
	return false
}
