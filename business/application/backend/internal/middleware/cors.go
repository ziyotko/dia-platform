package middleware

import (
	"net/http"
	"strings"

	"application/config"

	"github.com/gin-gonic/gin"
)

// CORS 仅放行 config.yaml 中 allowed_origins 白名单内的来源。
// 原实现固定下发 Access-Control-Allow-Origin: *（任何站点都能在用户浏览器里调用本服务）；
// 与 portal / member 统一为白名单后，未命中来源不下发该响应头（同源部署不受影响）。
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && isAllowedOrigin(origin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-Request-Timestamp, X-Request-Nonce, X-Request-Signature")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Disposition")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
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
