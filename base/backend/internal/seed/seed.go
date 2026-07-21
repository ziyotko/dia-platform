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

	menus := []models.Menu{
		{AppCode: "base", Name: "控制台", Path: "/index", Component: "base/dashboard/index.vue", Type: "menu", Icon: "HomeFilled", Sort: 1, Status: 1},
		{AppCode: "base", Name: "系统管理", Path: "/system", Type: "directory", Icon: "Setting", Sort: 100, Status: 1},
		{AppCode: "base", Name: "租户管理", Path: "/system/tenant", Component: "base/tenant/index.vue", Type: "menu", Sort: 1, Status: 1},
		{AppCode: "base", Name: "应用管理", Path: "/system/app", Component: "base/app/index.vue", Type: "menu", Sort: 2, Status: 1},
		{AppCode: "base", Name: "用户管理", Path: "/system/user", Component: "base/user/index.vue", Type: "menu", Sort: 3, Status: 1},
		{AppCode: "base", Name: "角色管理", Path: "/system/role", Component: "base/role/index.vue", Type: "menu", Sort: 4, Status: 1},
		{AppCode: "base", Name: "菜单管理", Path: "/system/menu", Component: "base/menu/index.vue", Type: "menu", Sort: 5, Status: 1},
		{AppCode: "base", Name: "系统设置", Path: "/system/setting", Component: "base/setting/index.vue", Type: "menu", Sort: 6, Status: 1},
	}

	// 设置目录父ID
	var systemDir models.Menu
	for i := range menus {
		if err := db.DB.Create(&menus[i]).Error; err != nil {
			return err
		}
		if menus[i].Name == "系统管理" {
			systemDir = menus[i]
		}
	}

	// 将系统管理子菜单挂到目录下
	for i := range menus {
		if menus[i].Type == "menu" && menus[i].Sort >= 1 && menus[i].Sort <= 6 && menus[i].Name != "控制台" {
			menus[i].ParentID = systemDir.ID
			if err := db.DB.Model(&menus[i]).Update("parent_id", systemDir.ID).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
