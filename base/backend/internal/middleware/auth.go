package middleware

import (
	"strings"

	"base/internal/models"
	"base/internal/service"
	"base/pkg/db"
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

		// 吊销校验：登出的 token 进黑名单；改密等场景让该用户旧 token 全部失效。
		// Redis 读失败时不再放行（fail-closed），但返回业务码 1 而非 401：缓存故障只让本次请求
		// 失败并提示稍后重试，不会把用户登出；只有确实已失效才返回 401。
		tokenSvc := service.TokenService{}
		revoked, err := tokenSvc.IsTokenRevoked(claims.ID)
		if err != nil {
			response.Fail(c, "登录状态校验失败，请稍后重试")
			c.Abort()
			return
		}
		if revoked {
			response.FailWithCode(c, response.CodeUnauthorized, tokenRevokedMessage)
			c.Abort()
			return
		}
		if claims.IssuedAt != nil {
			before, ok, err := tokenSvc.UserRevokedBefore(claims.UserID)
			if err != nil {
				response.Fail(c, "登录状态校验失败，请稍后重试")
				c.Abort()
				return
			}
			if ok && claims.IssuedAt.Time.Before(before) {
				response.FailWithCode(c, response.CodeUnauthorized, tokenRevokedMessage)
				c.Abort()
				return
			}
		}

		// 账号存活校验（不依赖 Redis）：被禁用 / 已删除的账号立即失效；
		// 并把 tenant_id 改成以数据库为准——用户被调到其它租户后，旧 token 里的租户快照
		// 不能继续当访问凭据（否则会按旧租户读到数据）。base_user 按主键查询，开销可忽略。
		var user models.User
		if err := db.DB.Select("id", "tenant_id", "status", "password_changed_at").
			First(&user, claims.UserID).Error; err != nil {
			response.FailWithCode(c, response.CodeUnauthorized, "账号不存在或已注销，请重新登录")
			c.Abort()
			return
		}
		if user.Status != 1 {
			response.FailWithCode(c, response.CodeUnauthorized, "账号已被禁用，请联系管理员")
			c.Abort()
			return
		}
		// 改密后旧 token 立即失效（落库口径；Redis 里的 revoked-before 丢失也不会漏）
		if claims.IssuedAt != nil && models.IsTokenStale(claims.IssuedAt.Time, user.PasswordChangedAt) {
			response.FailWithCode(c, response.CodeUnauthorized, tokenRevokedMessage)
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("tenantID", user.TenantID)
		c.Set("tokenID", claims.ID)
		if claims.ExpiresAt != nil {
			c.Set("tokenExp", claims.ExpiresAt.Time)
		}
		c.Next()
	}
}
