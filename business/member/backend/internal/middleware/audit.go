package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"member/internal/models"
	"member/pkg/db"

	"github.com/gin-gonic/gin"
)

// OperationLog records admin write operations (POST/PUT/DELETE) to the
// member_operation_logs table for audit purposes. GET queries are intentionally
// skipped to avoid noisy logs.
func OperationLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			c.Next()
			return
		}

		start := time.Now()

		// 仅捕获 JSON / 表单请求体，跳过 multipart 文件上传，避免大体积与二进制内容入日志
		contentType := c.GetHeader("Content-Type")
		captureBody := strings.HasPrefix(contentType, "application/json") ||
			strings.HasPrefix(contentType, "application/x-www-form-urlencoded")

		var bodyBytes []byte
		if captureBody && c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		writer := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = writer

		c.Next()

		memberID := GetMemberID(c)
		username := GetUsername(c)

		// 截断过长内容，避免数据库膨胀
		params := truncate(string(bodyBytes), 2000)
		result := truncate(writer.body.String(), 2000)

		log := models.OperationLog{
			MemberID:    memberID,
			Username:    username,
			Module:      c.FullPath(),
			Action:      c.Request.Method + " " + c.Request.URL.Path,
			Method:      c.Request.Method,
			Path:        c.Request.URL.Path,
			IP:          c.ClientIP(),
			Params:      params,
			Result:      result,
			Status:      1,
			Duration:    time.Since(start).Milliseconds(),
			OperationAt: start,
		}
		if c.Writer.Status() >= 400 {
			log.Status = 0
		}
		// 异步写入，不阻塞请求
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

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
