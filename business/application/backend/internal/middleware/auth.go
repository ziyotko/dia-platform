package middleware

import (
	"strings"
	"time"

	"application/internal/models"
	"application/pkg/db"
	"application/pkg/jwt"
	"application/pkg/redis"
	"application/pkg/response"

	"github.com/gin-gonic/gin"
	jwtlib "github.com/golang-jwt/jwt/v5"
)

const (
	CtxUserID    = "userID"
	CtxUsername  = "username"
	CtxAdminID   = "adminID"
	CtxAdminName = "adminUsername"
	CtxRoleCode  = "roleCode"
	// CtxTokenID / CtxTokenExp：本次请求 Token 的 jti 与过期时间（登出写黑名单需要）
	CtxTokenID  = "tokenID"
	CtxTokenExp = "tokenExp"
)

// UserAuth validates applicant (frontend) JWT token
func UserAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		claims, err := jwt.ParseUserToken(token)
		if err != nil {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		// 登出黑名单：本人主动登出后，该 Token 在自然过期前一律作废
		if rejectIfRevoked(c, claims.ID) {
			return
		}
		// 回查账号：token 有效期内账号可能已被删除或禁用，只验签名会让它们继续用到过期。
		// 按主键查一列，开销可忽略。
		var user models.User
		if err := db.DB.Select("id", "status", "password_changed_at").First(&user, claims.UserID).Error; err != nil || user.Status != 1 {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		// 改密后旧 Token 立即失效（原先最长要等 JWT 自然过期，默认 24h）
		if rejectIfStale(c, claims.IssuedAt, user.PasswordChangedAt) {
			return
		}
		setTokenContext(c, claims.ID, claims.ExpiresAt)
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxUsername, claims.Username)
		c.Next()
	}
}

// AdminAuth validates admin JWT token (manager / reviewer / super admin)
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		claims, err := jwt.ParseAdminToken(token)
		if err != nil {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		// 角色码缺失时不能放行：PermissionGuard 会因查不到角色而 403，但未挂权限校验的
		// 管理接口（看板、上传等）会直接读到一个空角色。历史遗留的「只有 user_id 的 token」
		// 也会在这里被拦住。
		if claims.RoleCode == "" {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		// 登出黑名单：与管理端登录页的「退出登录」对应
		if rejectIfRevoked(c, claims.ID) {
			return
		}
		// 回查账号：禁用/删除后旧 token 不应继续生效；**角色码也以数据库为准**
		// （token 里的 role_code 是签发时的快照，若不回查，超管被降级为 manager 后
		//  旧 token 在 24h 有效期内仍会按超管放行）。
		var admin models.Admin
		if err := db.DB.Select("id", "status", "role_code", "password_changed_at").First(&admin, claims.AdminID).Error; err != nil || admin.Status != 1 {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		if admin.RoleCode == "" {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		// 改密后旧 Token 立即失效（本人改密或管理员在「账号管理」里改密都会写入该时间）
		if rejectIfStale(c, claims.IssuedAt, admin.PasswordChangedAt) {
			return
		}
		setTokenContext(c, claims.ID, claims.ExpiresAt)
		c.Set(CtxAdminID, claims.AdminID)
		c.Set(CtxAdminName, claims.Username)
		c.Set(CtxRoleCode, admin.RoleCode)
		c.Next()
	}
}

// rejectIfRevoked 检查登出黑名单；命中或 Redis 异常（fail-closed）时写响应并返回 true。
func rejectIfRevoked(c *gin.Context, jti string) bool {
	revoked, err := redis.IsTokenRevoked(jti)
	if err != nil {
		response.Fail(c, "服务暂时不可用，请稍后重试")
		c.Abort()
		return true
	}
	if revoked {
		response.FailWithCode(c, response.CodeUnauthorized, "已退出登录，请重新登录")
		c.Abort()
		return true
	}
	return false
}

// rejectIfStale 判断 Token 是否因「改密」而失效（iat 早于 password_changed_at）。
// 两个时间都是秒级精度：恰好等于失效点（同一秒内签发）视为有效。
func rejectIfStale(c *gin.Context, issuedAt *jwtlib.NumericDate, changedAt *time.Time) bool {
	if issuedAt == nil || !models.IsTokenStale(issuedAt.Time, changedAt) {
		return false
	}
	response.FailWithCode(c, response.CodeUnauthorized, "密码已修改，请重新登录")
	c.Abort()
	return true
}

// setTokenContext 记录 jti 与过期时间，供登出接口写黑名单（TTL = 剩余有效期）。
func setTokenContext(c *gin.Context, jti string, exp *jwtlib.NumericDate) {
	c.Set(CtxTokenID, jti)
	if exp != nil {
		c.Set(CtxTokenExp, exp.Time)
	}
}

// RoleGuard checks if admin has one of the required roles
func RoleGuard(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleCode, exists := c.Get(CtxRoleCode)
		if !exists {
			response.Forbidden(c)
			c.Abort()
			return
		}
		roleStr, _ := roleCode.(string)
		for _, r := range roles {
			if r == roleStr {
				c.Next()
				return
			}
		}
		response.Forbidden(c)
		c.Abort()
	}
}

// PermissionGuard checks if admin role has the required permission
func PermissionGuard(perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleCode, exists := c.Get(CtxRoleCode)
		if !exists {
			response.Forbidden(c)
			c.Abort()
			return
		}
		roleStr, _ := roleCode.(string)
		perms, ok := GetRolePermissions()[roleStr]
		if !ok {
			response.Forbidden(c)
			c.Abort()
			return
		}
		for _, p := range perms {
			if p == perm {
				c.Next()
				return
			}
		}
		response.Forbidden(c)
		c.Abort()
	}
}

// PermissionGuardAny 允许命中其中任意一个权限即放行。
// 用于多个角色通过各自不同权限到达的同一个接口（例如通用上传接口）。
func PermissionGuardAny(perms ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleCode, exists := c.Get(CtxRoleCode)
		if !exists {
			response.Forbidden(c)
			c.Abort()
			return
		}
		roleStr, _ := roleCode.(string)
		rolePerms, ok := GetRolePermissions()[roleStr]
		if !ok {
			response.Forbidden(c)
			c.Abort()
			return
		}
		for _, p := range rolePerms {
			for _, want := range perms {
				if p == want {
					c.Next()
					return
				}
			}
		}
		response.Forbidden(c)
		c.Abort()
	}
}

func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

// GetTokenID 本次请求 Token 的 jti（登出时写入黑名单）。
func GetTokenID(c *gin.Context) string {
	v, _ := c.Get(CtxTokenID)
	s, _ := v.(string)
	return s
}

// GetTokenExpiry 本次请求 Token 的过期时间；取不到时返回零值
// （黑名单 TTL 会用兜底值，不会写出永不过期的键）。
func GetTokenExpiry(c *gin.Context) time.Time {
	v, _ := c.Get(CtxTokenExp)
	t, _ := v.(time.Time)
	return t
}

// Context helpers
func GetUserID(c *gin.Context) uint64 {
	id, _ := c.Get(CtxUserID)
	v, _ := id.(uint64)
	return v
}

func GetUsername(c *gin.Context) string {
	name, _ := c.Get(CtxUsername)
	v, _ := name.(string)
	return v
}

func GetAdminID(c *gin.Context) uint64 {
	id, _ := c.Get(CtxAdminID)
	v, _ := id.(uint64)
	return v
}

func GetAdminUsername(c *gin.Context) string {
	name, _ := c.Get(CtxAdminName)
	v, _ := name.(string)
	return v
}

func GetRoleCode(c *gin.Context) string {
	code, _ := c.Get(CtxRoleCode)
	v, _ := code.(string)
	return v
}

// GetRolePermissions returns permission codes per role.
//
// ⚠️ 这是**生效的**权限表（鉴权只看这里）；数据库 `application_roles.permissions` 仅用于后台展示。
// 因此 seed 里的同一份权限必须与此处保持一致（见 internal/seed/seed.go 的 permMap），否则角色管理页显示错。
func GetRolePermissions() map[string][]string {
	return map[string][]string{
		"super_admin": {
			"dashboard:view",
			"batch:view", "batch:create", "batch:edit", "batch:delete", "batch:publish",
			"category:view", "category:create", "category:edit", "category:delete",
			"application:view", "application:preliminary",
			"review:assign", "review:score",
			"result:publish", "result:manage",
			"certificate:manage",
			"announcement:manage",
			"notification:send",
			"user:manage", "admin:manage", "role:manage",
			"expert:manage",
			"audit:view",
		},
		"manager": {
			"dashboard:view",
			"batch:view", "batch:create", "batch:edit", "batch:publish",
			"category:view", "category:create", "category:edit",
			"application:view", "application:preliminary",
			"review:assign",
			"result:publish", "result:manage",
			"certificate:manage",
			"announcement:manage",
			"notification:send",
			"user:manage",
			"expert:manage",
			"audit:view",
		},
		"reviewer": {
			"dashboard:view",
			// 评审人只做「我的评审」：/admin/reviews* 由 AdminAuth 保护并按 reviewer_id
			// 自限（MyAssignments 按本人过滤、GetAssignment 校验归属），不需要
			// application:view / batch:view —— 否则评审人可翻看全部申报与他人评分意见，
			// 既泄露申报人资料，又破坏评审独立性（与 database.md 的声明一致）。
			"review:score",
		},
	}
}
