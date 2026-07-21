package middleware

import (
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

func TenantGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, exists := c.Get("tenantID")
		if !exists || tenantID.(uint64) == 0 {
			// 允许超级管理员 tenantID=0 继续，后续可按需拦截
			// 此处仅做示例：要求必须有租户上下文
		}
		c.Next()
	}
}

func SuperAdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, _ := c.Get("tenantID")
		if tenantID == nil || tenantID.(uint64) != 0 {
			response.FailWithCode(c, response.CodeForbidden, "仅平台超级管理员可操作")
			c.Abort()
			return
		}
		c.Next()
	}
}
