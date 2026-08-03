package middleware

import (
	"strconv"
	"strings"

	"conference/pkg/jwt"
	"conference/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	CtxUserID      = "userID"
	CtxUsername    = "username"
	CtxMemberLevel = "memberLevel"
	CtxBranch      = "branch"
	CtxAdminID     = "adminID"
	CtxAdminName   = "adminUsername"
	CtxRoleCode    = "roleCode"
)

// MemberAuth validates member JWT token
func MemberAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		claims, err := jwt.ParseMemberToken(token)
		if err != nil {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxUsername, claims.Username)
		c.Set(CtxMemberLevel, claims.MemberLevel)
		c.Set(CtxBranch, claims.Branch)
		c.Next()
	}
}

// AdminAuth validates admin JWT token
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

// PermissionGuard checks if admin has the required permission
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

func GetMemberLevel(c *gin.Context) string {
	level, _ := c.Get(CtxMemberLevel)
	v, _ := level.(string)
	return v
}

func GetBranch(c *gin.Context) string {
	branch, _ := c.Get(CtxBranch)
	v, _ := branch.(string)
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

// GetRolePermissions returns cached permissions map
func GetRolePermissions() map[string][]string {
	return map[string][]string{
		"super_admin": {
			"meeting:view", "meeting:create", "meeting:edit", "meeting:delete", "meeting:close",
			"registration:view", "registration:approve", "registration:import", "registration:export",
			"signin:view", "signin:manage", "signin:export",
			"vote:view", "vote:create", "vote:result", "vote:export",
			"finance:view", "finance:manage", "finance:refund", "finance:export",
			"live:view", "live:manage",
			"survey:view", "survey:create", "survey:result", "survey:export",
			"credit:view", "credit:manage",
			"archive:view", "archive:export",
			"notification:send", "notification:view",
			"user:manage",
			"audit:view",
			"dashboard:view",
			"role:manage",
		},
		"meeting_admin": {
			"meeting:view", "meeting:create", "meeting:edit", "meeting:close",
			"registration:view", "registration:approve", "registration:import", "registration:export",
			"signin:view", "signin:manage", "signin:export",
			"live:view", "live:manage",
			"survey:view", "survey:create", "survey:result", "survey:export",
			"archive:view", "archive:export",
			"dashboard:view",
		},
		"finance": {
			"finance:view", "finance:manage", "finance:refund", "finance:export",
			"dashboard:view",
		},
		"branch_admin": {
			"meeting:view",
			"registration:view", "registration:approve",
			"signin:view",
			"vote:view", "vote:result",
			"finance:view",
			"live:view",
			"survey:view", "survey:result",
			"credit:view",
			"archive:view",
			"dashboard:view",
		},
		"supervisor": {
			"meeting:view",
			"registration:view", "registration:export",
			"signin:view", "signin:export",
			"vote:view", "vote:result",
			"finance:view", "finance:export",
			"live:view",
			"survey:view", "survey:result",
			"credit:view",
			"archive:view",
			"audit:view",
			"dashboard:view",
		},
	}
}

// Helper to parse uint from string
func ParseUint(s string) uint64 {
	v, _ := strconv.ParseUint(s, 10, 64)
	return v
}

func ParseIntDefault(s string, def int) int {
	v, err := strconv.Atoi(s)
	if err != nil || v <= 0 {
		return def
	}
	return v
}
