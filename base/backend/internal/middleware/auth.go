package middleware

import (
	"strings"

	"base/pkg/jwt"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			response.FailWithCode(c, response.CodeUnauthorized, "请先登录")
			c.Abort()
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.FailWithCode(c, response.CodeUnauthorized, "Token格式错误")
			c.Abort()
			return
		}
		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			response.FailWithCode(c, response.CodeUnauthorized, "Token无效或已过期")
			c.Abort()
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("tenantID", claims.TenantID)
		c.Next()
	}
}
