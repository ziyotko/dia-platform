package middleware

import (
	"strings"

	"application/internal/models"
	"application/pkg/db"
	"application/pkg/jwt"
	"application/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	CtxUserID    = "userID"
	CtxUsername  = "username"
	CtxAdminID   = "adminID"
	CtxAdminName = "adminUsername"
	CtxRoleCode  = "roleCode"
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
		// 回查账号：token 有效期内账号可能已被删除或禁用，只验签名会让它们继续用到过期。
		// 按主键查一列，开销可忽略。
		var user models.User
		if err := db.DB.Select("id", "status").First(&user, claims.UserID).Error; err != nil || user.Status != 1 {
			response.Unauthorized(c)
			c.Abort()
			return
		}
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
		// 回查账号：角色变更/禁用/删除后，旧 token 不应继续生效。
		var admin models.Admin
		if err := db.DB.Select("id", "status").First(&admin, claims.AdminID).Error; err != nil || admin.Status != 1 {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		c.Set(CtxAdminID, claims.AdminID)
		c.Set(CtxAdminName, claims.Username)
		c.Set(CtxRoleCode, claims.RoleCode)
		c.Next()
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

// GetRolePermissions returns permission codes per role
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
			"batch:view",
			"application:view",
			"review:score",
		},
	}
}
