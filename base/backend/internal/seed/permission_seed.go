package seed

import (
	"errors"

	"base/internal/models"
	"base/pkg/db"

	"gorm.io/gorm"
)

// permissionSeed 权限点种子。Code 全局唯一，作为幂等键。
// Type=menu 的为分组节点（不参与接口匹配，仅用于前端权限树展示）；Type=api 的参与
// middleware.PermissionAuth 的 method + path 匹配，Path 使用去除 /base/api/v1 前缀的相对路径。
type permissionSeed struct {
	Code       string
	Name       string
	Type       string
	ParentCode string
	Method     string
	Path       string
}

var basePermissionSeeds = []permissionSeed{
	// ---------- 用户管理 ----------
	{Code: "base:user", Name: "用户管理", Type: "menu"},
	{Code: "base:user:list", Name: "查看用户列表", Type: "api", ParentCode: "base:user", Method: "GET", Path: "/users"},
	{Code: "base:user:detail", Name: "查看用户详情", Type: "api", ParentCode: "base:user", Method: "GET", Path: "/users/:id"},
	{Code: "base:user:create", Name: "新增用户", Type: "api", ParentCode: "base:user", Method: "POST", Path: "/users"},
	{Code: "base:user:update", Name: "编辑用户", Type: "api", ParentCode: "base:user", Method: "PUT", Path: "/users/:id"},
	{Code: "base:user:delete", Name: "删除用户", Type: "api", ParentCode: "base:user", Method: "DELETE", Path: "/users/:id"},
	{Code: "base:user:assign-role", Name: "分配角色", Type: "api", ParentCode: "base:user", Method: "POST", Path: "/users/:id/roles"},
	{Code: "base:user:reset-password", Name: "重置密码", Type: "api", ParentCode: "base:user", Method: "POST", Path: "/users/:id/reset-password"},

	// ---------- 角色管理 ----------
	{Code: "base:role", Name: "角色管理", Type: "menu"},
	{Code: "base:role:list", Name: "查看角色列表", Type: "api", ParentCode: "base:role", Method: "GET", Path: "/roles"},
	{Code: "base:role:detail", Name: "查看角色详情", Type: "api", ParentCode: "base:role", Method: "GET", Path: "/roles/:id"},
	{Code: "base:role:create", Name: "新增角色", Type: "api", ParentCode: "base:role", Method: "POST", Path: "/roles"},
	{Code: "base:role:update", Name: "编辑角色", Type: "api", ParentCode: "base:role", Method: "PUT", Path: "/roles/:id"},
	{Code: "base:role:delete", Name: "删除角色", Type: "api", ParentCode: "base:role", Method: "DELETE", Path: "/roles/:id"},
	{Code: "base:role:assign-menu", Name: "分配菜单", Type: "api", ParentCode: "base:role", Method: "POST", Path: "/roles/:id/menus"},
	{Code: "base:role:assign-permission", Name: "分配权限", Type: "api", ParentCode: "base:role", Method: "POST", Path: "/roles/:id/permissions"},

	// ---------- 菜单管理 ----------
	{Code: "base:menu", Name: "菜单管理", Type: "menu"},
	{Code: "base:menu:tree", Name: "查看菜单树", Type: "api", ParentCode: "base:menu", Method: "GET", Path: "/menus/tree"},
	{Code: "base:menu:create", Name: "新增菜单", Type: "api", ParentCode: "base:menu", Method: "POST", Path: "/menus"},
	{Code: "base:menu:update", Name: "编辑菜单", Type: "api", ParentCode: "base:menu", Method: "PUT", Path: "/menus/:id"},
	{Code: "base:menu:delete", Name: "删除菜单", Type: "api", ParentCode: "base:menu", Method: "DELETE", Path: "/menus/:id"},

	// ---------- 接口权限 ----------
	{Code: "base:permission", Name: "接口权限", Type: "menu"},
	{Code: "base:permission:tree", Name: "查看权限树", Type: "api", ParentCode: "base:permission", Method: "GET", Path: "/permissions/tree"},
	{Code: "base:permission:create", Name: "新增权限", Type: "api", ParentCode: "base:permission", Method: "POST", Path: "/permissions"},
	{Code: "base:permission:update", Name: "编辑权限", Type: "api", ParentCode: "base:permission", Method: "PUT", Path: "/permissions/:id"},
	{Code: "base:permission:delete", Name: "删除权限", Type: "api", ParentCode: "base:permission", Method: "DELETE", Path: "/permissions/:id"},

	// ---------- 机构管理 ----------
	{Code: "base:org", Name: "机构管理", Type: "menu"},
	{Code: "base:org:tree", Name: "查看机构树", Type: "api", ParentCode: "base:org", Method: "GET", Path: "/organizations/tree"},
	{Code: "base:org:detail", Name: "查看机构详情", Type: "api", ParentCode: "base:org", Method: "GET", Path: "/organizations/:id"},
	{Code: "base:org:create", Name: "新增机构", Type: "api", ParentCode: "base:org", Method: "POST", Path: "/organizations"},
	{Code: "base:org:update", Name: "编辑机构", Type: "api", ParentCode: "base:org", Method: "PUT", Path: "/organizations/:id"},
	{Code: "base:org:delete", Name: "删除机构", Type: "api", ParentCode: "base:org", Method: "DELETE", Path: "/organizations/:id"},

	// ---------- 数据字典 ----------
	{Code: "base:dict", Name: "数据字典", Type: "menu"},
	{Code: "base:dict:list", Name: "查看字典列表", Type: "api", ParentCode: "base:dict", Method: "GET", Path: "/dicts"},
	{Code: "base:dict:detail", Name: "查看字典详情", Type: "api", ParentCode: "base:dict", Method: "GET", Path: "/dicts/:id"},
	{Code: "base:dict:by-code", Name: "按编码取字典", Type: "api", ParentCode: "base:dict", Method: "GET", Path: "/dicts/code/:code"},
	{Code: "base:dict:create", Name: "新增字典", Type: "api", ParentCode: "base:dict", Method: "POST", Path: "/dicts"},
	{Code: "base:dict:update", Name: "编辑字典", Type: "api", ParentCode: "base:dict", Method: "PUT", Path: "/dicts/:id"},
	{Code: "base:dict:delete", Name: "删除字典", Type: "api", ParentCode: "base:dict", Method: "DELETE", Path: "/dicts/:id"},
	{Code: "base:dict:save-items", Name: "保存字典项", Type: "api", ParentCode: "base:dict", Method: "POST", Path: "/dicts/:id/items"},

	// ---------- 应用管理 ----------
	{Code: "base:app", Name: "应用管理", Type: "menu"},
	{Code: "base:app:list", Name: "查看应用列表", Type: "api", ParentCode: "base:app", Method: "GET", Path: "/apps"},
	{Code: "base:app:detail", Name: "查看应用详情", Type: "api", ParentCode: "base:app", Method: "GET", Path: "/apps/:id"},
	{Code: "base:app:create", Name: "新增应用", Type: "api", ParentCode: "base:app", Method: "POST", Path: "/apps"},
	{Code: "base:app:update", Name: "编辑应用", Type: "api", ParentCode: "base:app", Method: "PUT", Path: "/apps/:id"},
	{Code: "base:app:delete", Name: "删除应用", Type: "api", ParentCode: "base:app", Method: "DELETE", Path: "/apps/:id"},

	// ---------- 应用实例 ----------
	{Code: "base:app-instance", Name: "应用实例", Type: "menu"},
	{Code: "base:app-instance:list", Name: "查看应用实例", Type: "api", ParentCode: "base:app-instance", Method: "GET", Path: "/app-instances"},
	{Code: "base:app-instance:create", Name: "开通应用", Type: "api", ParentCode: "base:app-instance", Method: "POST", Path: "/app-instances"},
	{Code: "base:app-instance:update", Name: "编辑应用实例", Type: "api", ParentCode: "base:app-instance", Method: "PUT", Path: "/app-instances/:id"},
	{Code: "base:app-instance:delete", Name: "删除应用实例", Type: "api", ParentCode: "base:app-instance", Method: "DELETE", Path: "/app-instances/:id"},

	// ---------- 消息管理 ----------
	{Code: "base:message", Name: "消息管理", Type: "menu"},
	{Code: "base:message:list", Name: "查看消息", Type: "api", ParentCode: "base:message", Method: "GET", Path: "/messages"},
	{Code: "base:message:detail", Name: "查看消息详情", Type: "api", ParentCode: "base:message", Method: "GET", Path: "/messages/:id"},
	{Code: "base:message:create", Name: "新建草稿", Type: "api", ParentCode: "base:message", Method: "POST", Path: "/messages"},
	{Code: "base:message:update", Name: "编辑草稿", Type: "api", ParentCode: "base:message", Method: "PUT", Path: "/messages/:id"},
	{Code: "base:message:send", Name: "发送消息", Type: "api", ParentCode: "base:message", Method: "POST", Path: "/messages/send"},
	{Code: "base:message:send-draft", Name: "发送草稿", Type: "api", ParentCode: "base:message", Method: "POST", Path: "/messages/:id/send"},
	{Code: "base:message:read", Name: "标记已读", Type: "api", ParentCode: "base:message", Method: "POST", Path: "/messages/:id/read"},
	{Code: "base:message:read-all", Name: "全部已读", Type: "api", ParentCode: "base:message", Method: "POST", Path: "/messages/read-all"},
	{Code: "base:message:delete", Name: "删除消息", Type: "api", ParentCode: "base:message", Method: "DELETE", Path: "/messages/:id"},

	// ---------- 消息模板 ----------
	{Code: "base:message-template", Name: "消息模板", Type: "menu"},
	{Code: "base:message-template:list", Name: "查看模板列表", Type: "api", ParentCode: "base:message-template", Method: "GET", Path: "/message-templates"},
	{Code: "base:message-template:detail", Name: "查看模板详情", Type: "api", ParentCode: "base:message-template", Method: "GET", Path: "/message-templates/:id"},
	{Code: "base:message-template:create", Name: "新增模板", Type: "api", ParentCode: "base:message-template", Method: "POST", Path: "/message-templates"},
	{Code: "base:message-template:update", Name: "编辑模板", Type: "api", ParentCode: "base:message-template", Method: "PUT", Path: "/message-templates/:id"},
	{Code: "base:message-template:delete", Name: "删除模板", Type: "api", ParentCode: "base:message-template", Method: "DELETE", Path: "/message-templates/:id"},

	// ---------- 流程角色（工作流） ----------
	{Code: "base:workflow-role", Name: "流程角色", Type: "menu"},
	{Code: "base:workflow-role:list", Name: "查看流程角色列表", Type: "api", ParentCode: "base:workflow-role", Method: "GET", Path: "/workflow-roles"},
	{Code: "base:workflow-role:detail", Name: "查看流程角色详情", Type: "api", ParentCode: "base:workflow-role", Method: "GET", Path: "/workflow-roles/:id"},
	{Code: "base:workflow-role:create", Name: "新增流程角色", Type: "api", ParentCode: "base:workflow-role", Method: "POST", Path: "/workflow-roles"},
	{Code: "base:workflow-role:update", Name: "编辑流程角色", Type: "api", ParentCode: "base:workflow-role", Method: "PUT", Path: "/workflow-roles/:id"},
	{Code: "base:workflow-role:delete", Name: "删除流程角色", Type: "api", ParentCode: "base:workflow-role", Method: "DELETE", Path: "/workflow-roles/:id"},
	{Code: "base:workflow-role:assign-user", Name: "配置流程角色成员", Type: "api", ParentCode: "base:workflow-role", Method: "POST", Path: "/workflow-roles/:id/users"},
	{Code: "base:workflow-role:user-options", Name: "查看可选成员", Type: "api", ParentCode: "base:workflow-role", Method: "GET", Path: "/workflow-roles/user-options"},

	// ---------- 流程定义（工作流） ----------
	{Code: "base:workflow-def", Name: "流程定义", Type: "menu"},
	{Code: "base:workflow-def:list", Name: "查看流程列表", Type: "api", ParentCode: "base:workflow-def", Method: "GET", Path: "/workflows"},
	{Code: "base:workflow-def:detail", Name: "查看流程详情", Type: "api", ParentCode: "base:workflow-def", Method: "GET", Path: "/workflows/:id"},
	{Code: "base:workflow-def:create", Name: "新增流程", Type: "api", ParentCode: "base:workflow-def", Method: "POST", Path: "/workflows"},
	{Code: "base:workflow-def:update", Name: "编辑流程", Type: "api", ParentCode: "base:workflow-def", Method: "PUT", Path: "/workflows/:id"},
	{Code: "base:workflow-def:delete", Name: "删除流程", Type: "api", ParentCode: "base:workflow-def", Method: "DELETE", Path: "/workflows/:id"},
	{Code: "base:workflow-def:nodes", Name: "保存流程节点", Type: "api", ParentCode: "base:workflow-def", Method: "PUT", Path: "/workflows/:id/nodes"},
	{Code: "base:workflow-def:approver-options", Name: "查看审批人候选", Type: "api", ParentCode: "base:workflow-def", Method: "GET", Path: "/workflows/approver-options"},

	// ---------- 流程实例（工作流） ----------
	{Code: "base:workflow-instance", Name: "流程实例", Type: "menu"},
	{Code: "base:workflow-instance:list", Name: "查看流程实例", Type: "api", ParentCode: "base:workflow-instance", Method: "GET", Path: "/workflow-instances"},
	{Code: "base:workflow-instance:detail", Name: "查看流程实例详情", Type: "api", ParentCode: "base:workflow-instance", Method: "GET", Path: "/workflow-instances/:id"},
	{Code: "base:workflow-instance:start", Name: "发起流程", Type: "api", ParentCode: "base:workflow-instance", Method: "POST", Path: "/workflow-instances"},
	{Code: "base:workflow-instance:cancel", Name: "撤销流程", Type: "api", ParentCode: "base:workflow-instance", Method: "POST", Path: "/workflow-instances/:id/cancel"},
	{Code: "base:workflow-instance:delete", Name: "删除流程实例", Type: "api", ParentCode: "base:workflow-instance", Method: "DELETE", Path: "/workflow-instances/:id"},

	// ---------- 待办任务（工作流） ----------
	{Code: "base:workflow-task", Name: "待办任务", Type: "menu"},
	{Code: "base:workflow-task:list", Name: "查看我的待办", Type: "api", ParentCode: "base:workflow-task", Method: "GET", Path: "/workflow-tasks"},
	{Code: "base:workflow-task:approve", Name: "审批通过", Type: "api", ParentCode: "base:workflow-task", Method: "POST", Path: "/workflow-tasks/:id/approve"},
	{Code: "base:workflow-task:reject", Name: "审批驳回", Type: "api", ParentCode: "base:workflow-task", Method: "POST", Path: "/workflow-tasks/:id/reject"},

	// ---------- 操作日志 ----------
	{Code: "base:operation-log", Name: "操作日志", Type: "menu"},
	{Code: "base:operation-log:list", Name: "查看操作日志", Type: "api", ParentCode: "base:operation-log", Method: "GET", Path: "/operation-logs"},
	{Code: "base:operation-log:export", Name: "导出操作日志", Type: "api", ParentCode: "base:operation-log", Method: "GET", Path: "/operation-logs/export"},
	{Code: "base:operation-log:delete", Name: "删除操作日志", Type: "api", ParentCode: "base:operation-log", Method: "POST", Path: "/operation-logs/delete"},
	{Code: "base:operation-log:clear", Name: "清理操作日志", Type: "api", ParentCode: "base:operation-log", Method: "POST", Path: "/operation-logs/clear"},

	// ---------- 登录日志 ----------
	{Code: "base:login-log", Name: "登录日志", Type: "menu"},
	{Code: "base:login-log:list", Name: "查看登录日志", Type: "api", ParentCode: "base:login-log", Method: "GET", Path: "/login-logs"},
	{Code: "base:login-log:export", Name: "导出登录日志", Type: "api", ParentCode: "base:login-log", Method: "GET", Path: "/login-logs/export"},
	{Code: "base:login-log:delete", Name: "删除登录日志", Type: "api", ParentCode: "base:login-log", Method: "POST", Path: "/login-logs/delete"},
	{Code: "base:login-log:clear", Name: "清理登录日志", Type: "api", ParentCode: "base:login-log", Method: "POST", Path: "/login-logs/clear"},

	// ---------- 文件管理 ----------
	{Code: "base:file", Name: "文件管理", Type: "menu"},
	{Code: "base:file:list", Name: "查看文件列表", Type: "api", ParentCode: "base:file", Method: "GET", Path: "/files"},
	{Code: "base:file:upload", Name: "上传文件", Type: "api", ParentCode: "base:file", Method: "POST", Path: "/files/upload"},
	{Code: "base:file:delete", Name: "删除文件", Type: "api", ParentCode: "base:file", Method: "DELETE", Path: "/files/:id"},

	// ---------- 系统设置 ----------
	{Code: "base:setting", Name: "系统设置", Type: "menu"},
	{Code: "base:setting:get", Name: "查看设置", Type: "api", ParentCode: "base:setting", Method: "GET", Path: "/settings"},
	{Code: "base:setting:save", Name: "保存设置", Type: "api", ParentCode: "base:setting", Method: "PUT", Path: "/settings"},
	{Code: "base:setting:test-email", Name: "邮件测试", Type: "api", ParentCode: "base:setting", Method: "POST", Path: "/settings/email/test"},
}

// seedBasePermissions 幂等地写入底座权限点（按 code 匹配，已存在则同步名称/类型/父子/方法/路径）。
func seedBasePermissions() error {
	ids := make(map[string]uint64, len(basePermissionSeeds))
	for _, s := range basePermissionSeeds {
		var parentID uint64
		if s.ParentCode != "" {
			parentID = ids[s.ParentCode]
		}

		var existing models.Permission
		err := db.DB.Where("code = ?", s.Code).First(&existing).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			perm := models.Permission{
				AppCode:  "base",
				Code:     s.Code,
				Name:     s.Name,
				Type:     s.Type,
				ParentID: parentID,
				Path:     s.Path,
				Method:   s.Method,
				Status:   1,
			}
			if err := db.DB.Create(&perm).Error; err != nil {
				return err
			}
			ids[s.Code] = perm.ID
		case err != nil:
			return err
		default:
			if err := db.DB.Model(&models.Permission{}).Where("id = ?", existing.ID).Updates(map[string]interface{}{
				"app_code":  "base",
				"name":      s.Name,
				"type":      s.Type,
				"parent_id": parentID,
				"path":      s.Path,
				"method":    s.Method,
			}).Error; err != nil {
				return err
			}
			ids[s.Code] = existing.ID
		}
	}
	return grantBasePermissionsToSuperAdmin()
}

// grantBasePermissionsToSuperAdmin 把全部底座权限点授予平台超级管理员角色，
// 保证升级后原有平台管理员不会因为新增权限点而被拦截。
func grantBasePermissionsToSuperAdmin() error {
	var role models.Role
	err := db.DB.Where("tenant_id = ? AND code = ?", 0, "super_admin").First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	var perms []models.Permission
	if err := db.DB.Where("app_code = ?", "base").Find(&perms).Error; err != nil {
		return err
	}
	if len(perms) == 0 {
		return nil
	}
	return db.DB.Model(&role).Association("Perms").Append(perms)
}
