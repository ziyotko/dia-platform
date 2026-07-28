package middleware

import (
	"member/pkg/jwt"
	"member/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	CtxMemberID = "memberID"
	CtxUsername = "username"
	CtxIsAdmin  = "isAdmin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "未登录或token已过期")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "token格式错误")
			c.Abort()
			return
		}

		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "token无效或已过期")
			c.Abort()
			return
		}

		c.Set(CtxMemberID, claims.MemberID)
		c.Set(CtxUsername, claims.Username)
		c.Set(CtxIsAdmin, claims.IsAdmin)
		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, _ := c.Get(CtxIsAdmin)
		if isAdmin == nil || !isAdmin.(bool) {
			response.Forbidden(c, "需要管理员权限")
			c.Abort()
			return
		}
		c.Next()
	}
}

func GetMemberID(c *gin.Context) uint64 {
	id, _ := c.Get(CtxMemberID)
	if id == nil {
		return 0
	}
	return id.(uint64)
}
