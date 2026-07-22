package middleware

import (
	"strings"

	"base/internal/models"
	"base/pkg/db"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

// 不需要接口级权限校验的白名单：基础个人信息、菜单、权限、改密、仪表盘、我的应用。
var permissionWhitelist = map[string][]string{
	"/base/api/v1/auth/info":            {"GET"},
	"/base/api/v1/auth/menus":           {"GET"},
	"/base/api/v1/auth/permissions":     {"GET"},
	"/base/api/v1/auth/change-password": {"POST"},
	"/base/api/v1/dashboard/stats":      {"GET"},
	"/base/api/v1/app-instances/my":     {"GET"},
}

// PermissionAuth 基于 base_permission 表的接口级权限校验中间件。
// 匹配规则：当前请求方法 + 路由模板（或去除 /base/api/v1 前缀后的路径）与权限表中的 method/path 匹配。
// 超级管理员（tenantID == 0）直接放行；白名单接口直接放行；权限表为空时放行以兼容历史数据。
func PermissionAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			response.FailWithCode(c, response.CodeUnauthorized, "请先登录")
			c.Abort()
			return
		}

		// 超级管理员直接放行
		tenantID, _ := c.Get("tenantID")
		if tid, ok := tenantID.(uint64); ok && tid == 0 {
			c.Next()
			return
		}

		requiredMethod := c.Request.Method
		requiredPath := c.FullPath()
		if requiredPath == "" {
			requiredPath = c.Request.URL.Path
		}

		// 白名单接口直接放行
		if methods, ok := permissionWhitelist[requiredPath]; ok {
			for _, m := range methods {
				if strings.EqualFold(m, requiredMethod) {
					c.Next()
					return
				}
			}
		}

		uid := userID.(uint64)

		// 同时支持完整路径和去除 /base/api/v1 前缀的相对路径
		candidatePaths := []string{requiredPath}
		if strings.HasPrefix(requiredPath, "/base/api/v1") {
			candidatePaths = append(candidatePaths, strings.TrimPrefix(requiredPath, "/base/api/v1"))
		}

		allowed, err := hasPermission(uid, requiredMethod, candidatePaths)
		if err != nil {
			response.FailWithCode(c, response.CodeError, "权限校验失败")
			c.Abort()
			return
		}

		if !allowed {
			response.FailWithCode(c, response.CodeForbidden, "无权限访问该接口")
			c.Abort()
			return
		}

		c.Next()
	}
}

// hasPermission 判断用户是否拥有任一候选路径的接口权限。
// 若该用户没有任何角色关联的有效权限记录，则视为未启用权限校验，直接放行（兼容历史数据）。
func hasPermission(userID uint64, method string, candidatePaths []string) (bool, error) {
	// 一次性查出该用户所有角色关联的有效权限
	var perms []models.Permission
	err := db.DB.
		Model(&models.Permission{}).
		Joins("JOIN base_role_permission ON base_role_permission.permission_id = base_permission.id").
		Joins("JOIN base_user_role ON base_user_role.role_id = base_role_permission.role_id").
		Where("base_user_role.user_id = ? AND base_permission.status = ?", userID, 1).
		Find(&perms).Error
	if err != nil {
		return false, err
	}

	// 未配置任何权限时放行，避免空权限表导致系统不可用
	if len(perms) == 0 {
		return true, nil
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
				return true, nil
			}
		}
	}
	return false, nil
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
