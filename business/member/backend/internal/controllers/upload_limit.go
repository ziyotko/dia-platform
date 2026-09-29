package controllers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// maxUploadOverhead 除文件本身外 multipart 还包含表单字段与边界，留 1MB 余量。
const maxUploadOverhead = 1 << 20

// limitMultipartBody 在解析 multipart 之前限制整个请求体大小。
//
// 必要性：gin 的 c.FormFile / c.PostForm 会先把整个请求体解析完（超出
// MaxMultipartMemory 的部分落到 /tmp 临时文件），之后才判断 file.Size 已经太晚——
// 匿名接口（/upload-public）或已登录用户都能据此写满磁盘。
// BodyLimitMiddleware 明确跳过 multipart，因此每个上传入口必须自己做这层限制。
//
// 超限后后续的 FormFile / ParseMultipartForm 会返回 *http.MaxBytesError。
func limitMultipartBody(c *gin.Context, maxBytes int64) {
	if c.Request == nil || c.Request.Body == nil {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes+maxUploadOverhead)
}

// isBodyTooLarge 判断读取 multipart 时的错误是否为「超过请求体上限」。
func isBodyTooLarge(err error) bool {
	if err == nil {
		return false
	}
	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) {
		return true
	}
	// 兼容 mime/multipart / net/http 在部分路径上对错误的文本包装。
	return strings.Contains(err.Error(), "request body too large")
}
