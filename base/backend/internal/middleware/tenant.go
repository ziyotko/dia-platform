package middleware

import (
	"base/internal/models"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

// SuperAdminOnly 仅平台超级管理员（models.PlatformTenantID）可访问。
// 平台内所有「超管专属」判定统一走 models.IsPlatformTenant，不再散落写 tenantID == 0。
func SuperAdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !models.IsPlatformTenant(c.GetUint64("tenantID")) {
			response.FailWithCode(c, response.CodeForbidden, "仅平台超级管理员可操作")
			c.Abort()
			return
		}
		c.Next()
	}
}
