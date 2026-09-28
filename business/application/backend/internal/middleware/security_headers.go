package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders 为所有响应添加基础安全头（与 portal / member 一致）：
//  1. X-Content-Type-Options: nosniff，禁止浏览器 MIME 嗅探；
//  2. X-Frame-Options / Referrer-Policy，降低点击劫持与 Referer 泄露风险。
//
// 注意：本项目上传文件**不做静态托管**（只能通过带鉴权的 GET /member/files、GET /admin/files 读取），
// 因此不需要 portal / member 那样的「上传目录 + 危险扩展名强制下载」分支。
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}
