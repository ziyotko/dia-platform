package models

import (
	"gorm.io/gorm"

	"server/utils"
)

// menuSeedItem 描述一条待初始化的菜单数据。
type menuSeedItem struct {
	Name      string
	Path      string
	Component string
	APIPrefix string
	Icon      string
	Type      string
	Sort      int
	Status    int
	Children  []menuSeedItem
}

// defaultMenus 系统内置默认菜单树。
// 与前端 src/views 目录一一对应，路径与前端路由约定保持一致：
//   - 目录使用分组路径（如 /content），菜单使用完整绝对路径（如 /content/article），
//     便于侧边栏高亮（default-active=route.path）与 router.push 直接跳转。
//   - 组件路径使用相对 views 的形式（如 content/article），由前端 loadComponent 自动补全。
var defaultMenus = []menuSeedItem{
	{
		Name: "管理首页", Path: "/dashboard", Component: "dashboard/index",
		Icon: "HomeFilled", Type: "menu", Sort: 0, Status: 1, APIPrefix: "/dashboard",
	},
	{
		Name: "内容管理", Path: "/content",
		Icon: "FolderOpened", Type: "directory", Sort: 1, Status: 1,
		Children: []menuSeedItem{
			{Name: "待审核", Path: "/content/pending-audits", Component: "system/pending-audits", Icon: "Bell", Type: "menu", Sort: 1, Status: 1, APIPrefix: "/articles/my-audits"},
			{Name: "图文管理", Path: "/content/article", Component: "content/article", Icon: "Document", Type: "menu", Sort: 2, Status: 1, APIPrefix: "/articles"},
			{Name: "广告管理", Path: "/content/ad", Component: "content/ad", Icon: "Picture", Type: "menu", Sort: 3, Status: 1, APIPrefix: "/ads"},
			{Name: "链接管理", Path: "/content/link", Component: "content/link", Icon: "Link", Type: "menu", Sort: 4, Status: 1, APIPrefix: "/links"},
			{Name: "模板管理", Path: "/content/template", Component: "content/template", Icon: "Tickets", Type: "menu", Sort: 5, Status: 1, APIPrefix: "/templates"},
			{Name: "栏目管理", Path: "/content/column", Component: "content/column", Icon: "Grid", Type: "menu", Sort: 6, Status: 1, APIPrefix: "/columns"},
			{Name: "分类管理", Path: "/content/category", Component: "content/category", Icon: "Folder", Type: "menu", Sort: 7, Status: 1, APIPrefix: "/categories"},
			{Name: "标签管理", Path: "/content/tag", Component: "content/tag", Icon: "PriceTag", Type: "menu", Sort: 8, Status: 1, APIPrefix: "/tags"},
			{Name: "静态化管理", Path: "/content/static", Component: "content/static", Icon: "Monitor", Type: "menu", Sort: 9, Status: 1, APIPrefix: "/static"},
		},
	},
	{
		Name: "数据统计", Path: "/statistics",
		Icon: "DataAnalysis", Type: "directory", Sort: 2, Status: 1,
		Children: []menuSeedItem{
			{Name: "内容数据", Path: "/statistics/analytics", Component: "statistics/analytics", Icon: "TrendCharts", Type: "menu", Sort: 1, Status: 1, APIPrefix: "/analytics/article-trend"},
			{Name: "文章统计", Path: "/statistics/article", Component: "statistics/article", Icon: "DocumentChecked", Type: "menu", Sort: 2, Status: 1, APIPrefix: "/articles/author-stats"},
			{Name: "分类统计", Path: "/statistics/category", Component: "statistics/category", Icon: "PieChart", Type: "menu", Sort: 3, Status: 1, APIPrefix: "/categories/stats"},
			{Name: "标签统计", Path: "/statistics/label", Component: "statistics/label", Icon: "Histogram", Type: "menu", Sort: 4, Status: 1, APIPrefix: "/tags/stats"},
		},
	},
	{
		Name: "系统配置", Path: "/system",
		Icon: "Operation", Type: "directory", Sort: 3, Status: 1,
		Children: []menuSeedItem{
			{Name: "用户管理", Path: "/system/users", Component: "system/users", Icon: "User", Type: "menu", Sort: 1, Status: 1, APIPrefix: "/users"},
			{Name: "部门管理", Path: "/system/departments", Component: "system/departments", Icon: "School", Type: "menu", Sort: 2, Status: 1, APIPrefix: "/departments"},
			{Name: "机构管理", Path: "/system/orgs", Component: "system/orgs", Icon: "OfficeBuilding", Type: "menu", Sort: 3, Status: 1, APIPrefix: "/organizations"},
			{Name: "角色管理", Path: "/system/roles", Component: "system/roles", Icon: "Avatar", Type: "menu", Sort: 4, Status: 1, APIPrefix: "/roles"},
			{Name: "菜单管理", Path: "/system/menus", Component: "system/menus", Icon: "Menu", Type: "menu", Sort: 5, Status: 1, APIPrefix: "/menus"},
			{Name: "流程角色", Path: "/system/workflow-roles", Component: "system/workflow-roles", Icon: "Stamp", Type: "menu", Sort: 6, Status: 1, APIPrefix: "/workflow-roles"},
			{Name: "流程管理", Path: "/system/workflows", Component: "system/workflows", Icon: "SetUp", Type: "menu", Sort: 7, Status: 1, APIPrefix: "/workflows"},
			{Name: "操作日志", Path: "/system/logs", Component: "system/logs", Icon: "Memo", Type: "menu", Sort: 8, Status: 1, APIPrefix: "/logs"},
			{Name: "登录日志", Path: "/system/login-logs", Component: "system/login-logs", Icon: "Key", Type: "menu", Sort: 9, Status: 1, APIPrefix: "/login-logs"},
		},
	},
	{
		Name: "基础配置", Path: "/config",
		Icon: "Setting", Type: "directory", Sort: 4, Status: 1,
		Children: []menuSeedItem{
			{Name: "静态化设置", Path: "/staticization", Component: "staticization/index", Icon: "Cpu", Type: "menu", Sort: 1, Status: 1, APIPrefix: "/settings"},
			{Name: "系统设置", Path: "/settings", Component: "settings/index", Icon: "Tools", Type: "menu", Sort: 2, Status: 1, APIPrefix: "/settings"},
		},
	},
}

// SeedDefaultMenus 启动时初始化系统默认菜单树，缺失时自动创建（幂等，按 parent_id + name 判定）。
func SeedDefaultMenus() {
	for i := range defaultMenus {
		seedMenu(0, &defaultMenus[i])
	}
}

// seedMenu 递归插入菜单：若 (parent_id, name) 已存在则跳过，否则创建，然后递归处理子菜单。
func seedMenu(parentID uint, item *menuSeedItem) {
	var existing Menu
	err := utils.DB.Where("parent_id = ? AND name = ?", parentID, item.Name).First(&existing).Error
	if err == nil {
		// 已存在，仅递归确保其子菜单存在
		for i := range item.Children {
			seedMenu(existing.ID, &item.Children[i])
		}
		return
	}
	if err != gorm.ErrRecordNotFound {
		utils.Logger.Warnf("检查默认菜单[%s]失败: %v", item.Name, err)
		return
	}
	menu := Menu{
		ParentID:  parentID,
		Name:      item.Name,
		Path:      item.Path,
		Component: item.Component,
		APIPrefix: item.APIPrefix,
		Icon:      item.Icon,
		Type:      item.Type,
		Sort:      item.Sort,
		Status:    item.Status,
	}
	if createErr := utils.DB.Create(&menu).Error; createErr != nil {
		utils.Logger.Warnf("创建默认菜单[%s]失败: %v", item.Name, createErr)
		return
	}
	utils.Logger.Infof("已自动创建默认菜单: %s", item.Name)
	for i := range item.Children {
		seedMenu(menu.ID, &item.Children[i])
	}
}
