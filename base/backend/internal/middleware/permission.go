package middleware

import (
	"net/http"
	"strings"

	"base/internal/models"
	"base/pkg/db"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

// 不需要接口级权限校验的白名单：基础个人信息、菜单、权限、改密、仪表盘、我的应用、未读消息数。
var permissionWhitelist = map[string][]string{
	"/base/api/v1/auth/info":             {"GET"},
	"/base/api/v1/auth/menus":            {"GET"},
	"/base/api/v1/auth/permissions":      {"GET"},
	"/base/api/v1/auth/change-password":  {"POST"},
	"/base/api/v1/dashboard/stats":       {"GET"},
	"/base/api/v1/app-instances/my":      {"GET"},
	"/base/api/v1/messages/unread-count": {"GET"},
}

// PermissionAuth 基于 base_permission 表的接口级权限校验中间件。
// 匹配规则：当前请求方法 + 路由模板（或去除 /base/api/v1 前缀后的路径）与权限表中的 method/path 匹配。
//
// 放行策略（自上而下）：
//  1. 平台超级管理员（`models.IsPlatformTenant`）直接放行；
//  2. 租户管理员（base_user.is_admin）在本租户内直接放行，与「管理员可见全部菜单」保持同一口径；
//  3. 白名单接口直接放行；
//  4. 未分配任何权限的普通用户：仅放行 GET，写操作拒绝（不再整体放行）。
func PermissionAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			response.FailWithCode(c, response.CodeUnauthorized, "请先登录")
			c.Abort()
			return
		}

		// 平台超级管理员直接放行
		if models.IsPlatformTenant(c.GetUint64("tenantID")) {
			c.Next()
			return
		}

		requiredMethod := c.Request.Method
		requiredPath := c.FullPath()
		if requiredPath == "" {
			requiredPath = c.Request.URL.Path
		}

		// 白名单接口直接放行（放在租户管理员判定之前，避免高频接口多一次查库）
		if methods, ok := permissionWhitelist[requiredPath]; ok {
			for _, m := range methods {
				if strings.EqualFold(m, requiredMethod) {
					c.Next()
					return
				}
			}
		}

		uid := userID.(uint64)

		// 租户管理员直接放行（租户隔离仍由 service 层的 tenant_id 条件保证）
		if isTenantAdmin(uid) {
			c.Next()
			return
		}

		perms, err := userPermissions(uid)
		if err != nil {
			response.FailWithCode(c, response.CodeError, "权限校验失败")
			c.Abort()
			return
		}

		// 未分配任何接口权限：只读放行，写操作拒绝
		if len(perms) == 0 {
			if requiredMethod == http.MethodGet {
				c.Next()
				return
			}
			response.FailWithCode(c, response.CodeForbidden, "未分配该操作的接口权限，请联系管理员")
			c.Abort()
			return
		}

		if !matchPermission(perms, requiredMethod, requiredPath) {
			response.FailWithCode(c, response.CodeForbidden, "无权限访问该接口")
			c.Abort()
			return
		}

		c.Next()
	}
}

// isTenantAdmin 判断用户是否为租户管理员（base_user.is_admin）。
func isTenantAdmin(userID uint64) bool {
	var user models.User
	if err := db.DB.Select("id", "is_admin").First(&user, userID).Error; err != nil {
		return false
	}
	return user.IsAdmin
}

// userPermissions 一次性查出该用户所有角色关联的有效接口权限。
func userPermissions(userID uint64) ([]models.Permission, error) {
	var perms []models.Permission
	err := db.DB.
		Model(&models.Permission{}).
		Joins("JOIN base_role_permission ON base_role_permission.permission_id = base_permission.id").
		Joins("JOIN base_user_role ON base_user_role.role_id = base_role_permission.role_id").
		Where("base_user_role.user_id = ? AND base_permission.status = ?", userID, 1).
		Find(&perms).Error
	return perms, err
}

// matchPermission 判断权限集合是否覆盖请求方法 + 路径。
// 同时支持完整路径（/base/api/v1/users）和相对路径（/users）两种配置方式。
func matchPermission(perms []models.Permission, method, requiredPath string) bool {
	candidatePaths := []string{requiredPath}
	if strings.HasPrefix(requiredPath, "/base/api/v1") {
		candidatePaths = append(candidatePaths, strings.TrimPrefix(requiredPath, "/base/api/v1"))
	}

	for _, p := range perms {
		if p.Method == "" || p.Path == "" {
			continue
		}
		if !strings.EqualFold(p.Method, method) {
			continue
		}
		for _, path := range candidatePaths {
			if matchPath(p.Path, path) {
				return true
			}
		}
	}
	return false
}

// matchPath 支持 :param 通配符的简单路径匹配。
func matchPath(permPath, requestPath string) bool {
	permPath = normalizePath(permPath)
	requestPath = normalizePath(requestPath)

	permParts := strings.Split(permPath, "/")
	reqParts := strings.Split(requestPath, "/")

	if len(permParts) != len(reqParts) {
		return false
	}

	for i := range permParts {
		if strings.HasPrefix(permParts[i], ":") || strings.HasPrefix(permParts[i], "*") {
			continue
		}
		if permParts[i] != reqParts[i] {
			return false
		}
	}
	return true
}

func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.TrimSuffix(path, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}
