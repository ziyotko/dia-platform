package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"server/utils"
)

func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		protocol := c.Request.Proto

		c.Next()

		endTime := time.Now()
		latency := endTime.Sub(startTime)
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		logFields := logrus.Fields{
			"timestamp":   endTime.Format("2006-01-02 15:04:05"),
			"method":      method,
			"path":        path,
			"query":       query,
			"status_code": statusCode,
			"latency":     latency.String(),
			"client_ip":   clientIP,
			"user_agent":  userAgent,
			"protocol":    protocol,
		}

		if errorMessage != "" {
			logFields["error"] = errorMessage
		}

		if statusCode >= 500 {
			utils.Logger.WithFields(logFields).Error(fmt.Sprintf("%s %s", method, path))
		} else if statusCode >= 400 {
			utils.Logger.WithFields(logFields).Warn(fmt.Sprintf("%s %s", method, path))
		} else {
			utils.Logger.WithFields(logFields).Info(fmt.Sprintf("%s %s", method, path))
		}
	}
}
