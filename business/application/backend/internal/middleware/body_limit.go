package middleware

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"application/pkg/response"
	"application/pkg/utils"

	"github.com/gin-gonic/gin"
)

// defaultMaxJSONBodyMB 非 multipart 请求体的默认上限（MB）。与 portal / member 一致。
const defaultMaxJSONBodyMB = 64

// BodyLimitMiddleware 限制非 multipart 请求体大小（可通过 server.max_json_body_mb 调整）。
//
// 只针对非 multipart：文件上传由 upload_controller 限制 50MB 并由 Nginx 兜底；
// 而 JSON 请求体会被审计/日志类中间件整体读入内存，无上限时单个超大 body 即可打满进程内存。
func BodyLimitMiddleware(maxMB int) gin.HandlerFunc {
	if maxMB <= 0 {
		maxMB = defaultMaxJSONBodyMB
	}
	maxBytes := int64(maxMB) << 20

	reject := func(c *gin.Context) {
		utils.LogWarn("请求体过大（上限 %dMB）: %s %s from %s",
			maxMB, c.Request.Method, c.Request.URL.Path, c.ClientIP())
		response.Fail(c, fmt.Sprintf("请求体过大，已超过 %dMB 上限", maxMB))
		c.Abort()
	}

	return func(c *gin.Context) {
		if strings.HasPrefix(c.ContentType(), "multipart/form-data") {
			c.Next()
			return
		}

		if cl := c.Request.ContentLength; cl > maxBytes {
			reject(c)
			return
		} else if cl < 0 {
			body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBytes+1))
			if err != nil {
				utils.LogWarn("读取请求体失败: %s %s: %s", c.Request.Method, c.Request.URL.Path, err)
				response.Fail(c, "读取请求体失败")
				c.Abort()
				return
			}
			if int64(len(body)) > maxBytes {
				reject(c)
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
		}

		c.Next()
	}
}
