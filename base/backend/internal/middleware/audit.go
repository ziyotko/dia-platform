package middleware

import (
	"bytes"
	"io"
	"strings"
	"time"

	"base/internal/models"
	"base/pkg/db"
	"base/pkg/permmatch"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// auditBodyLimit 请求体/响应体落库长度上限：避免大响应把审计表撑爆
const auditBodyLimit = 4096

// maskedBody 凡写入明文密码的接口，请求体一律不记入日志（改为占位符）
const maskedBody = "[敏感参数已脱敏]"

// isSensitiveBody 判断该请求的 body 是否包含密码等敏感信息。
// 注意：这里只针对会携带密码的写接口，其余接口仍保留完整请求体便于追溯。
func isSensitiveBody(path string) bool {
	switch {
	case path == permmatch.APIPrefix()+"/auth/change-password":
		return true
	case path == permmatch.APIPrefix()+"/settings": // SMTP 密码等配置项
		return true
	case strings.HasPrefix(path, permmatch.APIPrefix()+"/users/") && strings.HasSuffix(path, "/reset-password"):
		return true
	}
	return false
}

func truncateBody(s string) string {
	if len(s) <= auditBodyLimit {
		return s
	}
	return s[:auditBodyLimit] + "...(已截断)"
}

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

		params := string(bodyBytes)
		if isSensitiveBody(c.Request.URL.Path) {
			params = maskedBody
		}

		log := models.OperationLog{
			TenantID:    tid,
			UserID:      uid,
			Username:    uname,
			Module:      c.FullPath(),
			Action:      c.Request.Method + " " + c.Request.URL.Path,
			Method:      c.Request.Method,
			Path:        c.Request.URL.Path,
			IP:          c.ClientIP(),
			Params:      truncateBody(params),
			Result:      truncateBody(writer.body.String()),
			Status:      1,
			Duration:    time.Since(start).Milliseconds(),
			OperationAt: start,
		}
		if c.Writer.Status() >= 400 {
			log.Status = 0
		}
		// 同步写入：异步 goroutine 在进程退出/重启时会丢失审计日志
		if err := db.DB.Create(&log).Error; err != nil {
			logrus.WithError(err).Warn("写入操作日志失败")
		}
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
