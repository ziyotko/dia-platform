package middleware

import (
	"path/filepath"
	"strings"

	"member/config"

	"github.com/gin-gonic/gin"
)

// 可能被浏览器解析执行从而导致存储型 XSS 的扩展名，强制以附件下载。
var dangerousExts = map[string]bool{
	".html": true, ".htm": true, ".shtml": true, ".svg": true,
	".js": true, ".mjs": true, ".xml": true, ".xhtml": true,
	".swf": true, ".php": true, ".jsp": true, ".asp": true,
	".aspx": true, ".hta": true,
	// 压缩包一律按附件下载（上传白名单已放行 zip/rar 用于多页扫描件）
	".zip": true, ".rar": true,
}

// SecureUploads 为上传静态资源（<upload_dir_prefix>/uploads）添加安全响应头：
// 1. X-Content-Type-Options: nosniff，禁止浏览器 MIME 嗅探；
// 2. 对危险类型强制 Content-Disposition: attachment，以附件下载而非内联渲染。
func SecureUploads() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, config.Cfg.Server.UploadDirPrefix+"/uploads") {
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
