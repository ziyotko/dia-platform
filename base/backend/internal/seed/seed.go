package seed

import (
	"errors"

	"base/internal/models"
	"base/pkg/db"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Run 初始化默认数据：底座菜单、底座权限点。
// 注意：平台超级管理员由 service.AuthService.EnsureSuperAdmin 创建（唯一实现），
// 需在调用 Run 之前完成——seedSuperAdminRole 会把菜单授予该 admin 用户。
func Run() error {
	if err := cleanupObsoleteMenus(); err != nil {
		return err
	}
	if err := seedBaseMenus(); err != nil {
		return err
	}
	if err := seedBasePermissions(); err != nil {
		return err
	}
	return nil
}

// obsoleteMenuComponents 历史遗留的占位菜单组件路径：这些页面文件已删除，
// 但旧版本的种子已写入数据库，保留会让用户点击后落到 404，因此启动时幂等清理。
var obsoleteMenuComponents = []string{
	"base/workflow/model/index.vue",
	"base/workflow/instance/index.vue",
	"base/workflow/task/index.vue",
	"base/workflow/designer/index.vue",
}

// cleanupObsoleteMenus 物理删除废弃菜单，并解除其与角色的菜单关联。
func cleanupObsoleteMenus() error {
	var ids []uint64
	if err := db.DB.Model(&models.Menu{}).Where("component IN ?", obsoleteMenuComponents).Pluck("id", &ids).Error; err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	// 先清关联表，避免角色菜单里残留指向已删除菜单的记录
	if err := db.DB.Exec("DELETE FROM base_role_menu WHERE menu_id IN ?", ids).Error; err != nil {
		return err
	}
	if err := db.DB.Delete(&models.Menu{}, ids).Error; err != nil {
		return err
	}
	logrus.Infof("已清理 %d 个历史遗留的占位菜单", len(ids))
	return nil
}

// baseMenuSeeds 底座默认菜单。
// ParentPath 为空表示顶层菜单；ParentPath 为父菜单的 Path。
// 注意：菜单为「按 (app_code, parent_id, name) 幂等补种」——已存在的不覆盖（保留管理员的改名/调整），
// 只补齐缺失项，因此后续新增菜单在已有环境重启后也会自动出现。
var baseMenuSeeds = []menuSeed{
	{AppCode: "base", Name: "控制台", Path: "/index", Component: "base/dashboard/index.vue", Type: "menu", Icon: "HomeFilled", Sort: 1},
	{AppCode: "base", Name: "系统管理", Path: "/system", Type: "directory", Icon: "Setting", Sort: 100},
	{AppCode: "base", Name: "消息管理", Path: "/message", Type: "directory", Icon: "Message", Sort: 200},
	{AppCode: "base", Name: "工作流管理", Path: "/workflow", Type: "directory", Icon: "Share", Sort: 300},

	{AppCode: "base", ParentPath: "/system", Name: "租户管理", Path: "/system/tenant", Component: "base/tenant/index.vue", Type: "menu", Sort: 1},
	{AppCode: "base", ParentPath: "/system", Name: "应用管理", Path: "/system/app", Component: "base/app/index.vue", Type: "menu", Sort: 2},
	{AppCode: "base", ParentPath: "/system", Name: "应用实例", Path: "/system/app-instance", Component: "base/app-instance/index.vue", Type: "menu", Sort: 3},
	{AppCode: "base", ParentPath: "/system", Name: "用户管理", Path: "/system/user", Component: "base/user/index.vue", Type: "menu", Sort: 4},
	{AppCode: "base", ParentPath: "/system", Name: "机构管理", Path: "/system/organization", Component: "base/organization/index.vue", Type: "menu", Sort: 5},
	{AppCode: "base", ParentPath: "/system", Name: "角色管理", Path: "/system/role", Component: "base/role/index.vue", Type: "menu", Sort: 6},
	{AppCode: "base", ParentPath: "/system", Name: "菜单管理", Path: "/system/menu", Component: "base/menu/index.vue", Type: "menu", Sort: 7},
	{AppCode: "base", ParentPath: "/system", Name: "权限管理", Path: "/system/permission", Component: "base/permission/index.vue", Type: "menu", Sort: 8},
	{AppCode: "base", ParentPath: "/system", Name: "审计日志", Path: "/system/log", Component: "base/log/index.vue", Type: "menu", Sort: 9},
	{AppCode: "base", ParentPath: "/system", Name: "登录日志", Path: "/system/login-log", Component: "base/login-log/index.vue", Type: "menu", Sort: 10},
	{AppCode: "base", ParentPath: "/system", Name: "数据字典", Path: "/system/dict", Component: "base/dict/index.vue", Type: "menu", Sort: 11},
	{AppCode: "base", ParentPath: "/system", Name: "文件管理", Path: "/system/file", Component: "base/file/index.vue", Type: "menu", Sort: 12},
	{AppCode: "base", ParentPath: "/system", Name: "系统设置", Path: "/system/setting", Component: "base/setting/index.vue", Type: "menu", Sort: 13},

	{AppCode: "base", ParentPath: "/message", Name: "消息列表", Path: "/message/list", Component: "base/message/index.vue", Type: "menu", Sort: 1},
	{AppCode: "base", ParentPath: "/message", Name: "消息模板", Path: "/message/template", Component: "base/message-template/index.vue", Type: "menu", Sort: 2},

	{AppCode: "base", ParentPath: "/workflow", Name: "流程角色", Path: "/workflow/role", Component: "base/workflow-role/index.vue", Type: "menu", Sort: 1},
	{AppCode: "base", ParentPath: "/workflow", Name: "流程定义", Path: "/workflow/definition", Component: "base/workflow/definition/index.vue", Type: "menu", Sort: 2},
	{AppCode: "base", ParentPath: "/workflow", Name: "流程实例", Path: "/workflow/instance", Component: "base/workflow/instance/index.vue", Type: "menu", Sort: 3},
	{AppCode: "base", ParentPath: "/workflow", Name: "我的待办", Path: "/workflow/task", Component: "base/workflow/task/index.vue", Type: "menu", Sort: 4},
}

type menuSeed struct {
	AppCode    string
	ParentPath string
	Name       string
	Path       string
	Component  string
	Type       string
	Icon       string
	Sort       int
	Target     string
}

func seedBaseMenus() error {
	// path -> id，用于解析父菜单
	idByPath := make(map[string]uint64)
	var created []models.Menu

	for _, s := range baseMenuSeeds {
		parentID := uint64(0)
		if s.ParentPath != "" {
			parentID = idByPath[s.ParentPath]
		}

		var existing models.Menu
		err := db.DB.Where("app_code = ? AND parent_id = ? AND name = ?", s.AppCode, parentID, s.Name).First(&existing).Error
		if err == nil {
			// 已存在：保留管理员的自定义，仅记录 id 供子菜单解析父级
			idByPath[s.Path] = existing.ID
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		menu := models.Menu{
			TenantID:  models.PlatformTenantID,
			AppCode:   s.AppCode,
			ParentID:  parentID,
			Name:      s.Name,
			Path:      s.Path,
			Component: s.Component,
			Type:      s.Type,
			Icon:      s.Icon,
			Sort:      s.Sort,
			Status:    1,
			Target:    s.Target,
		}
		if err := db.DB.Create(&menu).Error; err != nil {
			return err
		}
		idByPath[s.Path] = menu.ID
		created = append(created, menu)
	}

	return seedSuperAdminRole(created)
}

// seedSuperAdminRole 保证平台内置超级管理员角色存在，并把本次新增的菜单授予它。
// 只追加新增菜单，避免把管理员主动取消的菜单又加回来。
func seedSuperAdminRole(newMenus []models.Menu) error {
	var role models.Role
	err := db.DB.Where("tenant_id = ? AND code = ?", models.PlatformTenantID, "super_admin").First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		role = models.Role{
			TenantID: models.PlatformTenantID,
			Code:     "super_admin",
			Name:     "超级管理员",
			Status:   1,
		}
		if err := db.DB.Create(&role).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	if len(newMenus) > 0 {
		if err := db.DB.Model(&role).Association("Menus").Append(newMenus); err != nil {
			return err
		}
	}

	// 将角色赋给 admin 用户（幂等）
	var admin models.User
	if err := db.DB.Where("username = ? AND tenant_id = ?", "admin", models.PlatformTenantID).First(&admin).Error; err != nil {
		return err
	}
	return db.DB.Model(&admin).Association("Roles").Append(&role)
}
