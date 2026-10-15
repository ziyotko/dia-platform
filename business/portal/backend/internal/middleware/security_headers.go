package middleware

import (
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"portal/config"
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
// 2. 对危险类型强制 Content-Disposition: attachment，以附件下载而非内联渲染；
// 3. Content-Security-Policy / Strict-Transport-Security（见下方注释）。
//
// 说明：本服务只返回 JSON 与 /uploads 静态文件，SPA 的 index.html 由 Nginx 托管，
// 因此这里的 CSP 主要覆盖「被直接打开的上传文件/接口响应」，SPA 文档的 CSP 需在
// Nginx 上配置（见 DEPLOY.md「Nginx 反向代理示例」）。两层叠加形成纵深防御：
// script-src 'self' 会拦掉内联 <script> 与 on* 事件属性（存储型 XSS 的主要执行方式）。
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		// 不设 default-src：图片/媒体/样式等资源不受影响，仅限制脚本来源，
		// 避免误伤「富文本内联 base64 图片」「<video> 加载私有签名地址」等既有功能。
		c.Header("Content-Security-Policy", "script-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'self'")
		// HSTS 仅在 HTTPS 下由浏览器生效（HTTP 访问会被忽略），故只在请求确实走 HTTPS 时下发，
		// 避免内网纯 HTTP 部署被浏览器缓存住强制跳转策略。
		if c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// 上传目录实际挂载路径带全局前缀（如 /caamm/uploads），需与之匹配，否则危险类型强制下载失效
		uploadPrefix := config.AppConfig.Server.UploadDirPrefix + "/uploads"
		if strings.HasPrefix(c.Request.URL.Path, uploadPrefix) {
			ext := strings.ToLower(filepath.Ext(c.Request.URL.Path))
			if portalDangerousExts[ext] {
				c.Header("Content-Disposition", "attachment")
				c.Header("Content-Type", "application/octet-stream")
			}
		}
		c.Next()
	}
}
