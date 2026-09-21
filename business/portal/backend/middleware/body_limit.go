package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"server/utils"
)

// defaultMaxJSONBodyMB 非 multipart 请求体的默认上限（MB）。
// 取值需容纳「富文本正文内联 base64 图片」的文章保存请求（通常几 MB ~ 几十 MB），
// 同时挡住「单个请求携带超大 body 把进程内存打满」的放大攻击（此前无任何上限）。
const defaultMaxJSONBodyMB = 64

// BodyLimitMiddleware 限制非 multipart 请求体大小（上限可通过 server.max_json_body_mb 调整）。
//
// 为什么只针对非 multipart：
//   - 文件上传（multipart/form-data）由 upload_controller 按目录限制单文件大小
//     （视频 800MB、其它 50MB），并由 Nginx client_max_body_size 兜底，不能用 JSON 的上限去卡；
//   - 而 JSON 请求体会被防重放（replay_protection）与操作日志中间件整体读入内存（GetRawData），
//     无上限时任意已认证用户发一个超大 body 即可把进程内存打满。
//
// 实现：有 Content-Length 时直接比较（不读取 body，零开销）；chunked（无 Content-Length）时
// 用 LimitReader 多读 1 字节判定是否超限，再恢复 body 供后续中间件与控制器使用。
func BodyLimitMiddleware(maxMB int) gin.HandlerFunc {
	if maxMB <= 0 {
		maxMB = defaultMaxJSONBodyMB
	}
	maxBytes := int64(maxMB) << 20

	reject := func(c *gin.Context) {
		if utils.Logger != nil {
			utils.Logger.Warnf("请求体过大（上限 %dMB）: %s %s from %s",
				maxMB, c.Request.Method, c.Request.URL.Path, utils.RealIP(c))
		}
		c.JSON(http.StatusOK, utils.Error(1, fmt.Sprintf("请求体过大，已超过 %dMB 上限", maxMB)))
		c.Abort()
	}

	return func(c *gin.Context) {
		// 文件上传不受此限制（单文件大小由 upload_controller 与 Nginx 限制）
		if strings.HasPrefix(c.ContentType(), "multipart/form-data") {
			c.Next()
			return
		}

		if cl := c.Request.ContentLength; cl > maxBytes {
			reject(c)
			return
		} else if cl < 0 {
			// Content-Length 未知（分块传输）：以「上限 + 1 字节」为界读取，超出即拒绝
			body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBytes+1))
			if err != nil {
				utils.Logger.Warnf("读取请求体失败: %s %s: %s", c.Request.Method, c.Request.URL.Path, err)
				c.JSON(http.StatusOK, utils.Error(1, "读取请求体失败"))
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
