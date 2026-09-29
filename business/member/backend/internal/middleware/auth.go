package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	goredis "github.com/go-redis/redis/v8"

	"member/internal/models"
	"member/pkg/db"
	"member/pkg/jwt"
	"member/pkg/redis"
	"member/pkg/response"
)

const (
	CtxMemberID = "memberID"
	CtxUsername = "username"
	CtxIsAdmin  = "isAdmin"
	// CtxStatus 会员会籍状态（AdminOnly 用它判断管理员账号是否已停用）
	CtxStatus = "memberStatus"
	// CtxClaims 本次请求的 JWT 声明（登出需要 jti 与过期时间）
	CtxClaims = "memberClaims"

	// logoutBlacklistPrefix 登出黑名单键前缀。值固定写 "1"（只判存在性，不存 Token 原文，
	// 避免「能读 Redis 即等于拿到可用凭证」）。
	logoutBlacklistPrefix = "blacklist:"
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

		// 1) 登出黑名单：本人主动登出后，该 Token 在自然过期前一律作废。
		//    Redis 异常时按「已失效」处理（fail-closed），避免缓存故障变成鉴权放开。
		if _, err := redis.CaptchaClient.Get(redis.Ctx, logoutBlacklistPrefix+claims.ID).Result(); err == nil {
			response.Unauthorized(c, "已退出登录，请重新登录")
			c.Abort()
			return
		} else if err != goredis.Nil {
			response.ServerError(c, "服务暂时不可用，请稍后重试")
			c.Abort()
			return
		}

		// 2) 账号存活校验：被删除的账号立即失效（原先只验签名，删号后旧 Token 仍可用满 24h）。
		//    member_users 按主键查询，代价可忽略。
		var member models.Member
		if err := db.DB.Select("id", "username", "is_admin", "status",
			"password_changed_at", "token_invalid_before").First(&member, claims.MemberID).Error; err != nil {
			response.Unauthorized(c, "账号不存在或已注销，请重新登录")
			c.Abort()
			return
		}

		// 3) 改密 / 管理员变更后，旧 Token 立即失效（原先最长要等 JWT 自然过期，默认 24h）：
		//    password_changed_at = 本人改密、管理员重置密码；token_invalid_before = 管理员变更会籍状态。
		//    两个时间都由写入方截断到秒，与 JWT iat 的秒级精度对齐。
		// iat 缺失时无法判断「改密/状态变更」前后的先后关系，直接按不可信 Token 拒绝
		// （本系统签发的 Token 一定带 iat；缺失说明签发方异常）。
		if claims.IssuedAt == nil {
			response.Unauthorized(c, "token无效或已过期")
			c.Abort()
			return
		}
		if models.IsTokenStale(claims.IssuedAt.Time, member.PasswordChangedAt) {
			response.Unauthorized(c, "密码已修改，请重新登录")
			c.Abort()
			return
		}
		if models.IsTokenStale(claims.IssuedAt.Time, member.TokenInvalidBefore) {
			response.Unauthorized(c, "账号状态已变更，请重新登录")
			c.Abort()
			return
		}

		// 注意：这里不做 status 白名单拦截。member 的登录本身不限状态（注册中/待审核/
		// 待缴费/已过期会员都必须能登录完成入会与缴费流程），因此「禁用」由管理员动作
		// 写入 token_invalid_before 即时踢下线实现，而不是靠状态否决。

		c.Set(CtxMemberID, member.ID)
		c.Set(CtxUsername, member.Username)
		// 管理员标记以数据库为准（Token 里只是签发时的快照）：
		// 取消管理员身份后旧 Token 立即失去管理员权限，而不是等它自然过期。
		c.Set(CtxIsAdmin, member.IsAdmin)
		c.Set(CtxStatus, member.Status)
		c.Set(CtxClaims, claims)
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
		// 管理员账号被停用（会籍置为「已过期」）后立即失去后台权限：
		// 原先只看 is_admin，把一个管理员账号置为过期后，它重新登录仍是全权管理员，
		// 等于无法真正停用/降权（系统里也没有把 is_admin 置回 false 的入口）。
		if status, _ := c.Get(CtxStatus); status != models.MemberStatusActive {
			response.Forbidden(c, "管理员账号已停用，请联系超级管理员")
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

func GetUsername(c *gin.Context) string {
	name, _ := c.Get(CtxUsername)
	if name == nil {
		return ""
	}
	return name.(string)
}

// GetClaims 返回本次请求的 JWT 声明（由 Auth 中间件写入；登出需要 jti 与过期时间）。
func GetClaims(c *gin.Context) *jwt.MemberClaims {
	v, ok := c.Get(CtxClaims)
	if !ok {
		return nil
	}
	claims, _ := v.(*jwt.MemberClaims)
	return claims
}
