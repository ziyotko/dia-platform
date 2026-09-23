package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

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

		// 截断过长内容，避免数据库膨胀；请求体先做敏感字段脱敏（口令等不入库）
		params := paramsForLog(bodyBytes, operationLogParamsMax)
		result := truncate(writer.body.String(), operationLogParamsMax)

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

// 审计参数与响应的最大保存长度（按字符计）
const operationLogParamsMax = 2000

// 敏感字段名（小写比较）：写库前替换为 ***，
// 否则「新增会员」等接口提交的明文口令会直接落到 member_operation_logs.params。
var sensitiveKeys = map[string]bool{
	"password": true, "old_password": true, "new_password": true, "confirm_password": true,
	"repassword": true, "pwd": true, "token": true, "access_token": true,
	"secret": true, "authorization": true, "email_password": true, "smtp_password": true,
}

// urlencodedSensitive 非 JSON 体（application/x-www-form-urlencoded）里的敏感键值
var urlencodedSensitive = regexp.MustCompile(`(?i)\b(password|pwd|token|secret)=[^&]*`)

// truncate 按「字符（rune）」截断。
// 按字节切会切断 UTF-8 序列：中文请求体写 utf8mb4 列时报 1366，
// 而审计日志是 goroutine 异步写入、错误被丢弃，会导致审计记录静默丢失。
func truncate(s string, max int) string {
	if max <= 0 || utf8.RuneCountInString(s) <= max {
		return s
	}
	return string([]rune(s)[:max])
}

// paramsForLog 生成可入审计日志的请求参数：
// JSON 体做敏感字段递归脱敏后按 rune 截断；非 JSON 体退化为正则脱敏 + 截断。
func paramsForLog(body []byte, max int) string {
	if len(body) == 0 {
		return ""
	}
	var parsed any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return truncate(urlencodedSensitive.ReplaceAllString(string(body), "$1=***"), max)
	}
	redacted, err := json.Marshal(redactSensitive(parsed))
	if err != nil {
		return truncate(string(body), max)
	}
	return truncate(string(redacted), max)
}

// redactSensitive 递归地把 JSON 中的敏感字段值替换为 ***（数组与嵌套对象同样处理）。
func redactSensitive(v any) any {
	switch val := v.(type) {
	case map[string]any:
		for k, sub := range val {
			if sensitiveKeys[strings.ToLower(k)] {
				val[k] = "***"
				continue
			}
			val[k] = redactSensitive(sub)
		}
		return val
	case []any:
		for i, sub := range val {
			val[i] = redactSensitive(sub)
		}
		return val
	default:
		return v
	}
}
