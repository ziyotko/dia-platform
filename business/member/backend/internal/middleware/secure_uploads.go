package middleware

import (
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// 可能被浏览器解析执行从而导致存储型 XSS 的扩展名，强制以附件下载。
var dangerousExts = map[string]bool{
	".html": true, ".htm": true, ".shtml": true, ".svg": true,
	".js": true, ".mjs": true, ".xml": true, ".xhtml": true,
	".swf": true, ".php": true, ".jsp": true, ".asp": true,
	".aspx": true, ".hta": true,
}

// SecureUploads 为 /uploads 静态资源添加安全响应头：
// 1. X-Content-Type-Options: nosniff，禁止浏览器 MIME 嗅探；
// 2. 对危险类型强制 Content-Disposition: attachment，以附件下载而非内联渲染。
func SecureUploads() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/uploads") {
			c.Header("X-Content-Type-Options", "nosniff")
			ext := strings.ToLower(filepath.Ext(c.Request.URL.Path))
			if dangerousExts[ext] {
				c.Header("Content-Disposition", "attachment")
				c.Header("Content-Type", "application/octet-stream")
			}
		}
		c.Next()
	}
}
