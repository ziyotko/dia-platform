package middleware

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"member/pkg/response"
	"member/pkg/utils"

	"github.com/gin-gonic/gin"
)

// defaultMaxJSONBodyMB 非 multipart 请求体的默认上限（MB）。与 portal / application 一致。
const defaultMaxJSONBodyMB = 64

// BodyLimitMiddleware 限制非 multipart 请求体大小（可通过 server.max_json_body_mb 调整）。
//
// 只针对非 multipart：
//   - 文件上传由各上传入口在解析前限制（controllers/auth_controller.go 与 fee_controller.go 10MB、
//     charter_controller.go 20MB，见 controllers/upload_limit.go）并由 Nginx 兜底；
//   - 而 JSON 请求体会被操作日志中间件（middleware.OperationLog）整体读入内存，
//     无上限时任意已认证账号发一个超大 body 即可把进程内存打满。
//
// 实现：有 Content-Length 时直接比较（不读 body）；chunked（无 Content-Length）时用
// LimitReader 多读 1 字节判定是否超限，再还原 body 供后续中间件与控制器使用。
func BodyLimitMiddleware(maxMB int) gin.HandlerFunc {
	if maxMB <= 0 {
		maxMB = defaultMaxJSONBodyMB
	}
	maxBytes := int64(maxMB) << 20

	reject := func(c *gin.Context) {
		utils.LogWarn("请求体过大（上限 %dMB）: %s %s from %s",
			maxMB, c.Request.Method, c.Request.URL.Path, c.ClientIP())
		response.Error(c, response.CodeFail, fmt.Sprintf("请求体过大，已超过 %dMB 上限", maxMB))
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
				response.Error(c, response.CodeFail, "读取请求体失败")
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
