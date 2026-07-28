package middleware

import (
	"member/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		if query != "" {
			path = path + "?" + query
		}

		utils.Logger.WithFields(map[string]interface{}{
			"status":     statusCode,
			"method":     c.Request.Method,
			"path":       path,
			"ip":         c.ClientIP(),
			"latency":    latency.String(),
			"user_agent": c.Request.UserAgent(),
		}).Info("HTTP Request")
	}
}
