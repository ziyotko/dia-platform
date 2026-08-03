package middleware

import (
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// 可能被浏览器解析执行从而导致存储型 XSS 的扩展名，强制以附件下载。
var portalDangerousExts = map[string]bool{
	".html": true, ".htm": true, ".shtml": true, ".svg": true,
	".js": true, ".mjs": true, ".xml": true, ".xhtml": true,
	".swf": true, ".php": true, ".jsp": true, ".asp": true,
	".aspx": true, ".hta": true,
}

// SecurityHeaders 为响应添加基础安全头，并保护 /uploads 静态资源：
// 1. X-Content-Type-Options: nosniff，禁止 MIME 嗅探；
// 2. 对危险类型强制 Content-Disposition: attachment，以附件下载而非内联渲染。
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		if strings.HasPrefix(c.Request.URL.Path, "/uploads") {
			ext := strings.ToLower(filepath.Ext(c.Request.URL.Path))
			if portalDangerousExts[ext] {
				c.Header("Content-Disposition", "attachment")
				c.Header("Content-Type", "application/octet-stream")
			}
		}
		c.Next()
	}
}
