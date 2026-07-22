package seed

import (
	"base/internal/models"
	"base/pkg/db"
	"base/pkg/utils"
)

// Run 初始化默认数据：超级管理员、底座菜单
func Run() error {
	if err := seedSuperAdmin(); err != nil {
		return err
	}
	if err := seedBaseMenus(); err != nil {
		return err
	}
	return nil
}

func seedSuperAdmin() error {
	var count int64
	if err := db.DB.Model(&models.User{}).Where("tenant_id = ? AND username = ?", 0, "admin").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hash, err := utils.HashPassword("admin123")
	if err != nil {
		return err
	}
	return db.DB.Create(&models.User{
		TenantID: 0,
		Username: "admin",
		Password: hash,
		RealName: "超级管理员",
		Status:   1,
		IsAdmin:  true,
	}).Error
}

func seedBaseMenus() error {
	var count int64
	if err := db.DB.Model(&models.Menu{}).Where("app_code = ?", "base").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	dirs := []models.Menu{
		{AppCode: "base", Name: "控制台", Path: "/index", Component: "base/dashboard/index.vue", Type: "menu", Icon: "HomeFilled", Sort: 1, Status: 1},
		{AppCode: "base", Name: "系统管理", Path: "/system", Type: "directory", Icon: "Setting", Sort: 100, Status: 1},
		{AppCode: "base", Name: "消息管理", Path: "/message", Type: "directory", Icon: "Message", Sort: 200, Status: 1},
		{AppCode: "base", Name: "工作流管理", Path: "/workflow", Type: "directory", Icon: "Connection", Sort: 300, Status: 1},
	}

	dirIDs := make(map[string]uint64)
	for i := range dirs {
		if err := db.DB.Create(&dirs[i]).Error; err != nil {
			return err
		}
		dirIDs[dirs[i].Name] = dirs[i].ID
	}

	systemDirID := dirIDs["系统管理"]
	msgDirID := dirIDs["消息管理"]
	workflowDirID := dirIDs["工作流管理"]

	children := []models.Menu{
		{AppCode: "base", ParentID: systemDirID, Name: "租户管理", Path: "/system/tenant", Component: "base/tenant/index.vue", Type: "menu", Sort: 1, Status: 1},
		{AppCode: "base", ParentID: systemDirID, Name: "应用管理", Path: "/system/app", Component: "base/app/index.vue", Type: "menu", Sort: 2, Status: 1},
		{AppCode: "base", ParentID: systemDirID, Name: "用户管理", Path: "/system/user", Component: "base/user/index.vue", Type: "menu", Sort: 3, Status: 1},
		{AppCode: "base", ParentID: systemDirID, Name: "机构管理", Path: "/system/organization", Component: "base/organization/index.vue", Type: "menu", Sort: 4, Status: 1},
		{AppCode: "base", ParentID: systemDirID, Name: "角色管理", Path: "/system/role", Component: "base/role/index.vue", Type: "menu", Sort: 5, Status: 1},
		{AppCode: "base", ParentID: systemDirID, Name: "菜单管理", Path: "/system/menu", Component: "base/menu/index.vue", Type: "menu", Sort: 6, Status: 1},
		{AppCode: "base", ParentID: systemDirID, Name: "系统设置", Path: "/system/setting", Component: "base/setting/index.vue", Type: "menu", Sort: 7, Status: 1},

		{AppCode: "base", ParentID: msgDirID, Name: "消息列表", Path: "/message/list", Component: "base/message/index.vue", Type: "menu", Sort: 1, Status: 1},
		{AppCode: "base", ParentID: msgDirID, Name: "消息模板", Path: "/message/template", Component: "base/message-template/index.vue", Type: "menu", Sort: 2, Status: 1},

		{AppCode: "base", ParentID: workflowDirID, Name: "流程模型", Path: "/workflow/model", Component: "base/workflow/model/index.vue", Type: "menu", Sort: 1, Status: 1},
		{AppCode: "base", ParentID: workflowDirID, Name: "流程实例", Path: "/workflow/instance", Component: "base/workflow/instance/index.vue", Type: "menu", Sort: 2, Status: 1},
		{AppCode: "base", ParentID: workflowDirID, Name: "审批任务", Path: "/workflow/task", Component: "base/workflow/task/index.vue", Type: "menu", Sort: 3, Status: 1},
		{AppCode: "base", ParentID: workflowDirID, Name: "流程设计器", Path: "/workflow/designer", Component: "base/workflow/designer/index.vue", Type: "menu", Sort: 4, Status: 1},
	}

	for i := range children {
		if err := db.DB.Create(&children[i]).Error; err != nil {
			return err
		}
	}

	// 创建默认超级管理员角色并关联所有 base 菜单
	role := models.Role{
		TenantID: 0,
		Code:     "super_admin",
		Name:     "超级管理员",
		Status:   1,
	}
	if err := db.DB.Create(&role).Error; err != nil {
		return err
	}

	var allMenus []models.Menu
	if err := db.DB.Where("app_code = ?", "base").Find(&allMenus).Error; err != nil {
		return err
	}
	if err := db.DB.Model(&role).Association("Menus").Append(allMenus); err != nil {
		return err
	}

	// 将角色赋给 admin 用户
	var admin models.User
	if err := db.DB.Where("username = ? AND tenant_id = ?", "admin", 0).First(&admin).Error; err != nil {
		return err
	}
	if err := db.DB.Model(&admin).Association("Roles").Append(&role); err != nil {
		return err
	}

	return nil
}
