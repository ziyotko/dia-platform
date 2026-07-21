package adapter

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// BaseContext 底座平台注入的用户上下文
type BaseContext struct {
	UserID   uint64
	Username string
	TenantID uint64
}

const baseContextKey = "base_context"

// BaseAuthMiddleware 子应用 Gin 中间件示例
// 从请求头读取底座平台注入的用户信息
func BaseAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.GetHeader("X-Base-User-ID")
		username := c.GetHeader("X-Base-Username")
		tenantIDStr := c.GetHeader("X-Base-Tenant-ID")

		if userIDStr == "" || username == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少底座用户认证信息"})
			c.Abort()
			return
		}

		userID, _ := strconv.ParseUint(userIDStr, 10, 64)
		tenantID, _ := strconv.ParseUint(tenantIDStr, 10, 64)

		ctx := BaseContext{
			UserID:   userID,
			Username: username,
			TenantID: tenantID,
		}
		c.Set(baseContextKey, ctx)
		c.Next()
	}
}

// GetBaseContext 获取底座上下文
func GetBaseContext(c *gin.Context) (BaseContext, bool) {
	v, ok := c.Get(baseContextKey)
	if !ok {
		return BaseContext{}, false
	}
	ctx, ok := v.(BaseContext)
	return ctx, ok
}
