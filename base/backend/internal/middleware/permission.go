package middleware

import (
	"strings"

	"base/internal/models"
	"base/internal/service"
	"base/pkg/permmatch"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

// 不需要接口级权限校验的白名单：基础个人信息、菜单、权限、改密、仪表盘、我的应用、未读消息数，
// 以及工作流中「属于自己的数据」类接口（发起、我的申请、待办审批），
// 这些接口的归属校验在 service 层完成（只能操作自己发起或自己负责的任务）。
// 白名单键统一用**相对路径**（不含 APIPrefix，如 /auth/info）：
// 匹配时对请求路径取 permmatch.Candidates（完整路径 + 去前缀后的相对路径）逐个查表，
// 因此改前缀不需要动这里；写成完整路径也仍可命中。
var permissionWhitelist = map[string][]string{
	"/auth/info":             {"GET"},
	"/auth/menus":            {"GET"},
	"/auth/permissions":      {"GET"},
	"/auth/change-password":  {"POST"},
	"/auth/logout":           {"POST"},
	"/dashboard/stats":       {"GET"},
	"/app-instances/my":      {"GET"},
	"/messages/unread-count": {"GET"},

	// 工作流：发起流程需要先选流程定义
	"/workflows/options":               {"GET"},
	"/workflow-instances":              {"POST", "GET"},
	"/workflow-instances/:id":          {"GET"},
	"/workflow-instances/:id/cancel":   {"POST"},
	"/workflow-tasks":                  {"GET"},
	"/workflow-tasks/approver-options": {"GET"},
	"/workflow-tasks/:id/approve":      {"POST"},
	"/workflow-tasks/:id/reject":       {"POST"},
	"/workflow-tasks/:id/transfer":     {"POST"},
	"/workflow-tasks/:id/add-approver": {"POST"},
}

// inWhitelist 判断「方法 + 请求路径」是否命中白名单（请求路径的完整/相对形式都查）。
func inWhitelist(method, requestPath string) bool {
	for _, p := range permmatch.Candidates(requestPath) {
		methods, ok := permissionWhitelist[p]
		if !ok {
			continue
		}
		for _, m := range methods {
			if strings.EqualFold(m, method) {
				return true
			}
		}
	}
	return false
}

// PermissionAuth 基于 base_permission 表的接口级权限校验中间件。
// 匹配规则：当前请求方法 + 路由模板（或去除 permmatch.APIPrefix 前缀后的路径）与权限表中的 method/path 匹配。
//
// 放行策略（自上而下）：
//  1. 平台超级管理员（`models.IsPlatformTenant`）直接放行；
//  2. 租户管理员（base_user.is_admin）在本租户内直接放行，与「管理员可见全部菜单」保持同一口径；
//  3. 白名单接口直接放行；
//  4. 未分配任何权限的普通用户：除白名单外一律拒绝（只读接口同样需要显式授权）。
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
		if inWhitelist(requiredMethod, requiredPath) {
			c.Next()
			return
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

		// 未分配任何权限：一律拒绝（白名单接口已在上面放行）。
		// 不再「放行所有 GET」——那会让零权限用户读到 /settings（含 SMTP 密码）、/users、
		// /operation-logs 等敏感数据，只读接口也必须由角色显式授权。
		if len(perms) == 0 {
			response.FailWithCode(c, response.CodeForbidden, "未分配任何接口权限，请联系管理员在「角色管理 → 分配权限」中授权")
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
	return (service.UserService{}).IsAdmin(userID)
}

// userPermissions 一次性查出该用户所有角色关联的有效接口权限。
// 查询实现放在 service 层，与子应用代理入口的鉴权共用同一口径。
func userPermissions(userID uint64) ([]models.Permission, error) {
	return (service.PermissionService{}).UserPermissions(userID)
}

// matchPermission 判断权限集合是否覆盖请求方法 + 路径。
// 同时支持完整路径（permmatch.APIPrefix + /users）和相对路径（/users）两种配置方式；
// 路径匹配规则与子应用代理入口共用 pkg/permmatch。
func matchPermission(perms []models.Permission, method, requiredPath string) bool {
	candidatePaths := permmatch.Candidates(requiredPath)
	for _, p := range perms {
		if p.Method == "" || p.Path == "" {
			continue
		}
		if !strings.EqualFold(p.Method, method) {
			continue
		}
		for _, path := range candidatePaths {
			if permmatch.Match(p.Path, path) {
				return true
			}
		}
	}
	return false
}
