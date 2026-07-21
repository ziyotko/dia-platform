package middleware

import (
	"bytes"
	"io"
	"time"

	"base/internal/models"
	"base/pkg/db"

	"github.com/gin-gonic/gin"
)

func OperationLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		writer := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = writer

		c.Next()

		userID, _ := c.Get("userID")
		username, _ := c.Get("username")
		tenantID, _ := c.Get("tenantID")

		uid, _ := userID.(uint64)
		tid, _ := tenantID.(uint64)
		uname, _ := username.(string)

		log := models.OperationLog{
			TenantID:    tid,
			UserID:      uid,
			Username:    uname,
			Module:      c.FullPath(),
			Action:      c.Request.Method + " " + c.Request.URL.Path,
			Method:      c.Request.Method,
			Path:        c.Request.URL.Path,
			IP:          c.ClientIP(),
			Params:      string(bodyBytes),
			Result:      writer.body.String(),
			Status:      1,
			Duration:    time.Since(start).Milliseconds(),
			OperationAt: start,
		}
		if c.Writer.Status() >= 400 {
			log.Status = 0
		}
		// 异步写入
		go db.DB.Create(&log)
	}
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
