package middleware

import (
	"strings"

	"base/internal/service"
	"base/pkg/jwt"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

// tokenRevokedMessage 登出或被改密失效后统一提示，前端据此重新登录。
const tokenRevokedMessage = "登录状态已失效，请重新登录"

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

		// 吊销校验：登出的 token 进黑名单；改密等场景让该用户旧 token 全部失效
		tokenSvc := service.TokenService{}
		if tokenSvc.IsTokenRevoked(claims.ID) {
			response.FailWithCode(c, response.CodeUnauthorized, tokenRevokedMessage)
			c.Abort()
			return
		}
		if claims.IssuedAt != nil {
			if before, ok := tokenSvc.UserRevokedBefore(claims.UserID); ok && claims.IssuedAt.Time.Before(before) {
				response.FailWithCode(c, response.CodeUnauthorized, tokenRevokedMessage)
				c.Abort()
				return
			}
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("tenantID", claims.TenantID)
		c.Set("tokenID", claims.ID)
		if claims.ExpiresAt != nil {
			c.Set("tokenExp", claims.ExpiresAt.Time)
		}
		c.Next()
	}
}
