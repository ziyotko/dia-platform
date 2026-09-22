package models

import (
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"server/utils"
)

// 静态化管理页面会调用多个模块的接口（静态化任务/静态化日志/静态化服务监控），
// 因此该菜单声明多个 api_prefix（逗号分隔，见 middleware/api_prefix.go）。
// 另：该页面需读取全局设置（静态化输出路径 / 首页整体变灰，GET /settings 在 admin 组），故追加 /settings
// （静态化参数已并入「基础配置-系统设置」的「静态化设置」页签，与此处 /settings 同源，不再单设菜单）；
// 列表 Tab 还读 GET /columns/publishes 与 GET /articles/column-publishes（member 组），故追加 /columns、/articles。
const staticManagementAPIPrefix = "/static,/static-logs,/static-monitor,/settings,/columns,/articles"

// 「栏目管理」页面除栏目本身外，还需读取模板列表（页面层已合并进模板，栏目挂 template_id），
// 以及「栏目审核流程」下拉使用的全部流程列表（GET /workflows?all=1），
// 因此 api_prefix 追加 /templates 与 /workflows（/columns 已足够覆盖栏目读写）。
const columnManagementAPIPrefix = "/columns,/templates,/workflows"

// 「管理首页」除仪表盘数据外，还会拉取「待我审核」列表（GET /articles/my-audits，member 组）用于「待处理」卡片，
// 未同时授予「待审核」/「图文管理」菜单的角色否则会报「没有授权」并静默显示空列表。
const dashboardAPIPrefix = "/dashboard,/articles/my-audits"

// 以下为「页面实际调用的跨模块接口」补充（口径：菜单可见范围 = 接口可调用范围）：
//   - 「用户管理」用机构列表作「所属机构」下拉；
//   - 「部门管理」用机构树/机构成员（负责人候选按机构过滤）与全量用户列表；
//   - 「机构管理」用全量用户列表 + 其「内设机构」Tab 直接新增部门（POST /departments）；
//   - 「流程角色」用全量用户列表（成员选择）；
//   - 「流程管理」用全量用户与流程角色列表（节点审批人下拉）。
//
// 这些都是 admin 组的只读接口（除 GET /users 在 member 组），管理员默认拥有全部菜单不受影响；
// 只有「自定义角色只授部分菜单」时才会因缺少前缀报「没有授权」。
// 新增管理员页面时请同样检查：该页面调用的每个路径都要么被其菜单 api_prefix 覆盖，要么在豁免表里。
const userManagementAPIPrefix = "/users,/organizations"

const departmentManagementAPIPrefix = "/departments,/organizations,/users"

const organizationManagementAPIPrefix = "/organizations,/users,/departments"

const workflowManagementAPIPrefix = "/workflows,/users,/workflow-roles"

const workflowRoleManagementAPIPrefix = "/workflow-roles,/users"

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
// 与前端 src/views 目录一一对应（目录名 = 一级菜单语义），路径与前端路由约定保持一致：
//   - 目录使用分组路径（如 /content），菜单使用完整绝对路径（如 /content/article），
//     便于侧边栏高亮（default-active=route.path）与 router.push 直接跳转。
//   - 组件路径使用相对 views 的形式（如 content/article、config/static），由前端 loadComponent 自动补全。
//   - src/views 下的目录按功能模块划分：content（内容管理）/ statistics（数据统计）/ system（系统配置）/
//     config（基础配置：静态化管理 + 系统设置）/ dashboard / login / profile / layout / error。
//     新增页面时请放到对应模块目录下，不要把不同模块的页面混放在同一目录。
var defaultMenus = []menuSeedItem{
	{
		Name: "管理首页", Path: "/dashboard", Component: "dashboard/index",
		Icon: "HomeFilled", Type: "menu", Sort: 0, Status: 1, APIPrefix: dashboardAPIPrefix,
	},
	{
		Name: "内容管理", Path: "/content",
		Icon: "FolderOpened", Type: "directory", Sort: 1, Status: 1,
		Children: []menuSeedItem{
			{Name: "待审核", Path: "/content/pending-audits", Component: "content/pending-audits", Icon: "Bell", Type: "menu", Sort: 1, Status: 1, APIPrefix: "/articles/my-audits"},
			{Name: "图文管理", Path: "/content/article", Component: "content/article", Icon: "Document", Type: "menu", Sort: 2, Status: 1, APIPrefix: "/articles"},
			{Name: "广告管理", Path: "/content/ad", Component: "content/ad", Icon: "Picture", Type: "menu", Sort: 3, Status: 1, APIPrefix: "/ads"},
			{Name: "链接管理", Path: "/content/link", Component: "content/link", Icon: "Link", Type: "menu", Sort: 4, Status: 1, APIPrefix: "/links"},
			{Name: "模板管理", Path: "/content/template", Component: "content/template", Icon: "Tickets", Type: "menu", Sort: 5, Status: 1, APIPrefix: "/templates"},
			{Name: "栏目管理", Path: "/content/column", Component: "content/column", Icon: "Grid", Type: "menu", Sort: 6, Status: 1, APIPrefix: columnManagementAPIPrefix},
			{Name: "分类管理", Path: "/content/category", Component: "content/category", Icon: "Folder", Type: "menu", Sort: 7, Status: 1, APIPrefix: "/categories"},
			{Name: "标签管理", Path: "/content/tag", Component: "content/tag", Icon: "PriceTag", Type: "menu", Sort: 8, Status: 1, APIPrefix: "/tags"},
		},
	},
	{
		Name: "数据统计", Path: "/statistics",
		Icon: "DataAnalysis", Type: "directory", Sort: 2, Status: 1,
		Children: []menuSeedItem{
			{Name: "内容数据", Path: "/statistics/analytics", Component: "statistics/analytics", Icon: "TrendCharts", Type: "menu", Sort: 1, Status: 1, APIPrefix: "/analytics/article-trend"},
			{Name: "文章统计", Path: "/statistics/article", Component: "statistics/article", Icon: "DocumentChecked", Type: "menu", Sort: 2, Status: 1, APIPrefix: "/articles/author-stats"},
			{Name: "分类统计", Path: "/statistics/category", Component: "statistics/category", Icon: "PieChart", Type: "menu", Sort: 3, Status: 1, APIPrefix: "/categories/stats"},
			{Name: "标签统计", Path: "/statistics/tag", Component: "statistics/tag", Icon: "Histogram", Type: "menu", Sort: 4, Status: 1, APIPrefix: "/tags/stats"},
		},
	},
	{
		Name: "系统配置", Path: "/system",
		Icon: "Operation", Type: "directory", Sort: 3, Status: 1,
		Children: []menuSeedItem{
			{Name: "用户管理", Path: "/system/users", Component: "system/users", Icon: "User", Type: "menu", Sort: 1, Status: 1, APIPrefix: userManagementAPIPrefix},
			{Name: "部门管理", Path: "/system/departments", Component: "system/departments", Icon: "School", Type: "menu", Sort: 2, Status: 1, APIPrefix: departmentManagementAPIPrefix},
			{Name: "机构管理", Path: "/system/orgs", Component: "system/orgs", Icon: "OfficeBuilding", Type: "menu", Sort: 3, Status: 1, APIPrefix: organizationManagementAPIPrefix},
			{Name: "角色管理", Path: "/system/roles", Component: "system/roles", Icon: "Avatar", Type: "menu", Sort: 4, Status: 1, APIPrefix: "/roles"},
			{Name: "菜单管理", Path: "/system/menus", Component: "system/menus", Icon: "Menu", Type: "menu", Sort: 5, Status: 1, APIPrefix: "/menus"},
			{Name: "流程角色", Path: "/system/workflow-roles", Component: "system/workflow-roles", Icon: "Stamp", Type: "menu", Sort: 6, Status: 1, APIPrefix: workflowRoleManagementAPIPrefix},
			{Name: "流程管理", Path: "/system/workflows", Component: "system/workflows", Icon: "SetUp", Type: "menu", Sort: 7, Status: 1, APIPrefix: workflowManagementAPIPrefix},
			{Name: "操作日志", Path: "/system/logs", Component: "system/logs", Icon: "Memo", Type: "menu", Sort: 8, Status: 1, APIPrefix: "/logs"},
			{Name: "登录日志", Path: "/system/login-logs", Component: "system/login-logs", Icon: "Key", Type: "menu", Sort: 9, Status: 1, APIPrefix: "/login-logs"},
		},
	},
	{
		Name: "基础配置", Path: "/config",
		Icon: "Setting", Type: "directory", Sort: 4, Status: 1,
		Children: []menuSeedItem{
			// 静态化管理属于站点级运维操作（对应接口仅在管理员路由组），故与「系统设置」同放「基础配置」，
			// 避免误授予内容角色后出现「菜单可见、页面全报没有授权」
			{Name: "静态化管理", Path: "/config/static", Component: "config/static", Icon: "Monitor", Type: "menu", Sort: 0, Status: 1, APIPrefix: staticManagementAPIPrefix},
			{Name: "系统设置", Path: "/settings", Component: "config/settings", Icon: "Tools", Type: "menu", Sort: 1, Status: 1, APIPrefix: "/settings"},
		},
	},
}

// SeedDefaultMenus 启动时初始化系统默认菜单树，缺失时自动创建（幂等，按 parent_id + name 判定）。
func SeedDefaultMenus() {
	// 历史结构搬迁必须在播种之前执行：搬迁是「按名字找到旧行并改父级」，而播种是「按 parent_id + name 新建」。
	// 若先播种，老库中仍挂在旧目录下的内置菜单会在新目录下被新建一条，随后搬迁又把旧行搬过来，
	// 最终同 (parent_id, name) 出现两条（2026-09-14 日志实证：同一秒内 INSERT 静态化管理 + 已迁移菜单）。
	// 搬迁只依赖已存在的目录，因此全新安装是 no-op。
	relocateLegacyMenus()
	for i := range defaultMenus {
		seedMenu(0, &defaultMenus[i])
	}
	// 历史版本「静态化管理」只声明了 /static，导致 /static-logs、/static-monitor 不在授权范围内，
	// 这里对未自定义过该值的环境做一次幂等升级。
	upgradeMenuAPIPrefix("静态化管理", "/static", staticManagementAPIPrefix)
	upgradeMenuAPIPrefix("静态化管理", "/static,/static-logs,/static-monitor", staticManagementAPIPrefix)
	upgradeMenuAPIPrefix("静态化管理", "/static,/static-logs,/static-monitor,/settings", staticManagementAPIPrefix)
	// 历史版本「栏目管理」声明的是 /columns,/pages（页面层已移除）或早期仅 /columns，这里做幂等升级。
	upgradeMenuAPIPrefix("栏目管理", "/columns", columnManagementAPIPrefix)
	upgradeMenuAPIPrefix("栏目管理", "/columns,/pages", columnManagementAPIPrefix)
	upgradeMenuAPIPrefix("栏目管理", "/columns,/templates", columnManagementAPIPrefix)
	// 「管理首页」补 /articles/my-audits（供「待处理」卡片）。
	upgradeMenuAPIPrefix("管理首页", "/dashboard", dashboardAPIPrefix)
	// 历史版本各管理页只声明了自身前缀，缺少页面实际调用的跨模块只读接口（详见上方常量注释）。
	upgradeMenuAPIPrefix("用户管理", "/users", userManagementAPIPrefix)
	upgradeMenuAPIPrefix("部门管理", "/departments", departmentManagementAPIPrefix)
	upgradeMenuAPIPrefix("机构管理", "/organizations", organizationManagementAPIPrefix)
	upgradeMenuAPIPrefix("机构管理", "/organizations,/users", organizationManagementAPIPrefix)
	upgradeMenuAPIPrefix("流程管理", "/workflows", workflowManagementAPIPrefix)
	upgradeMenuAPIPrefix("流程角色", "/workflow-roles", workflowRoleManagementAPIPrefix)
	// 「静态化设置」独立页面已并入「系统设置」的「静态化设置」页签，清理旧环境残留菜单。
	removeLegacyStaticizationMenu()
	// 前端 src/views 目录按功能模块重排（2026-09-21）：待审核移入 content/、静态化管理与系统设置移入 config/、
	// 标签统计的 label 改名 tag，这里对未自行调整过组件的环境做一次幂等升级，避免菜单加载不到组件。
	upgradeMenuComponent("待审核", "system/pending-audits", "content/pending-audits")
	upgradeMenuComponent("静态化管理", "content/static", "config/static")
	upgradeMenuComponent("系统设置", "settings/index", "config/settings")
	upgradeMenuComponent("标签统计", "statistics/label", "statistics/tag")
	upgradeMenuPath("标签统计", "/statistics/label", "/statistics/tag")
	// 播种后再搬迁一次：覆盖「新父目录本次才被创建」的老库（搬迁要求目标目录已存在）。幂等。
	relocateLegacyMenus()
	// 兜底去重（幂等）：收敛历史版本「先播种后搬迁」在库中残留的重复内置菜单，并修正 role.permissions 引用。
	dedupeSeededMenus()
}

// relocateLegacyMenus 历史内置菜单的结构搬迁（幂等，只依赖已存在的目录）。
// 搬迁必须在 SeedDefaultMenus 的播种步骤之前执行，详见 SeedDefaultMenus 注释。
func relocateLegacyMenus() {
	// 历史版本「静态化管理」挂在「内容管理」下（非管理员即使被授权也调不通其接口），迁移到「基础配置」。
	moveMenuToParent("静态化管理", "内容管理", "基础配置", "/config/static")
}

// dedupeSeededMenus 幂等收敛内置菜单的重复行。
// 背景：早期版本的 SeedDefaultMenus 先播种后搬迁，老库中「静态化管理」会先在新父目录下被新建一条、
// 再把旧行搬过来，最终同 (parent_id, name, path, component) 出现两条 → 侧边栏/菜单管理/权限树重复。
// 处理：同「父级 + 名称 + 类型 + 路径 + 组件」只保留 id 最小的一条，其余在没有子菜单时删除，
// 并把引用了被删 ID 的 role.permissions 改指向保留的那条（否则该角色的菜单权限会静默丢失）。
func dedupeSeededMenus() {
	var menus []Menu
	if err := utils.DB.Order("id ASC").Find(&menus).Error; err != nil {
		utils.Logger.Warnf("菜单去重检查失败: %v", err)
		return
	}
	keptIDs := make(map[string]uint, len(menus))
	remapped := make(map[uint]uint)
	for i := range menus {
		menu := menus[i]
		key := fmt.Sprintf("%d|%s|%s|%s|%s", menu.ParentID, menu.Name, menu.Type, menu.Path, menu.Component)
		keepID, exists := keptIDs[key]
		if !exists {
			keptIDs[key] = menu.ID
			continue
		}
		// 仍有子菜单时不处理，避免子菜单变成孤儿
		var childCount int64
		if err := utils.DB.Model(&Menu{}).Where("parent_id = ?", menu.ID).Count(&childCount).Error; err != nil {
			continue
		}
		if childCount > 0 {
			continue
		}
		if err := utils.DB.Delete(&Menu{}, menu.ID).Error; err != nil {
			utils.Logger.Warnf("清理重复菜单[%s]失败: %v", menu.Name, err)
			continue
		}
		remapped[menu.ID] = keepID
		utils.Logger.Infof("已清理重复菜单: %s (删除 id=%d，保留 id=%d)", menu.Name, menu.ID, keepID)
	}
	if len(remapped) > 0 {
		remapRolePermissions(remapped)
	}
}

// remapRolePermissions 把 role.permissions 中对已删除菜单 ID 的引用改指向保留 ID（同时去重）。
func remapRolePermissions(remapped map[uint]uint) {
	var roles []Role
	if err := utils.DB.Find(&roles).Error; err != nil {
		utils.Logger.Warnf("修正角色菜单权限失败: %v", err)
		return
	}
	for i := range roles {
		role := roles[i]
		if strings.TrimSpace(role.Permissions) == "" {
			continue
		}
		parts := strings.Split(role.Permissions, ",")
		seen := make(map[uint]bool, len(parts))
		kept := make([]string, 0, len(parts))
		changed := false
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed == "" {
				continue
			}
			id64, err := strconv.ParseUint(trimmed, 10, 32)
			if err != nil {
				kept = append(kept, trimmed) // 非数字项（如历史遗留的 "*"）原样保留
				continue
			}
			id := uint(id64)
			if newID, ok := remapped[id]; ok {
				id = newID
				changed = true
			}
			if seen[id] {
				changed = true
				continue
			}
			seen[id] = true
			kept = append(kept, strconv.FormatUint(uint64(id), 10))
		}
		if !changed {
			continue
		}
		if err := utils.DB.Model(&Role{}).Where("id = ?", role.ID).
			Update("permissions", strings.Join(kept, ",")).Error; err != nil {
			utils.Logger.Warnf("修正角色[%s]菜单权限失败: %v", role.Code, err)
			continue
		}
		utils.Logger.Infof("已修正角色[%s]菜单权限中的重复菜单 ID", role.Code)
	}
}

// removeLegacyStaticizationMenu 幂等删除历史版本遗留的「静态化设置」独立菜单
// （该页已并入「基础配置-系统设置」的「静态化设置」页签，前端页面文件已删除）。
// 仅当菜单仍指向旧页面组件时才删除，避免误删使用方自行新增的同名菜单；
// 角色 permissions 中残留的菜单 ID 匹配不到已删除的菜单，无副作用。
func removeLegacyStaticizationMenu() {
	var menu Menu
	if err := utils.DB.Where("name = ? AND component = ?", "静态化设置", "staticization/index").First(&menu).Error; err != nil {
		return
	}
	if err := utils.DB.Delete(&menu).Error; err != nil {
		utils.Logger.Warnf("清理历史菜单[静态化设置]失败: %v", err)
		return
	}
	utils.Logger.Infof("已清理历史菜单: 静态化设置（已并入系统设置-静态化设置页签）")
}

// moveMenuToParent 幂等调整内置菜单的归属目录（仅当当前父目录仍为旧目录时），并同步 path。
func moveMenuToParent(menuName, oldParentName, newParentName, newPath string) {
	var menu Menu
	if err := utils.DB.Where("name = ?", menuName).First(&menu).Error; err != nil {
		return
	}
	var oldParent Menu
	if err := utils.DB.Where("name = ? AND type = ?", oldParentName, "directory").First(&oldParent).Error; err != nil {
		return
	}
	if menu.ParentID != oldParent.ID {
		return // 已被调整过或使用方自行归类，不覆盖
	}
	var newParent Menu
	if err := utils.DB.Where("name = ? AND type = ?", newParentName, "directory").First(&newParent).Error; err != nil {
		return
	}
	if err := utils.DB.Model(&menu).Updates(map[string]any{
		"parent_id": newParent.ID,
		"path":      newPath,
		"sort":      0, // 与全新安装的默认顺序保持一致（排在新目录首位）
	}).Error; err != nil {
		utils.Logger.Warnf("迁移菜单[%s]归属失败: %v", menuName, err)
		return
	}
	utils.Logger.Infof("已迁移菜单[%s]: %s -> %s (path: %s)", menuName, oldParentName, newParentName, newPath)
}

// upgradeMenuAPIPrefix 幂等升级内置菜单的 api_prefix：仅当当前值仍等于旧默认值时才更新，
// 避免覆盖使用方自行调整过的配置。
func upgradeMenuAPIPrefix(name, legacyPrefix, newPrefix string) {
	var menu Menu
	if err := utils.DB.Where("name = ? AND api_prefix = ?", name, legacyPrefix).First(&menu).Error; err != nil {
		return
	}
	if err := utils.DB.Model(&menu).Update("api_prefix", newPrefix).Error; err != nil {
		utils.Logger.Warnf("升级菜单[%s]接口前缀失败: %v", name, err)
		return
	}
	utils.Logger.Infof("已升级菜单[%s]接口前缀: %s -> %s", name, legacyPrefix, newPrefix)
}

// upgradeMenuComponent 幂等升级内置菜单的前端组件路径（口径同 upgradeMenuAPIPrefix：
// 仅当当前值仍等于旧默认值时才更新，避免覆盖使用方自行调整过的配置）。
// 前端 src/views 目录调整后，老库里残留的旧 component 会因找不到组件而被前端跳过（菜单点进去为空），
// 故每次目录重排都需在此登记一条升级。
func upgradeMenuComponent(name, legacyComponent, newComponent string) {
	var menu Menu
	if err := utils.DB.Where("name = ? AND component = ?", name, legacyComponent).First(&menu).Error; err != nil {
		return
	}
	if err := utils.DB.Model(&menu).Update("component", newComponent).Error; err != nil {
		utils.Logger.Warnf("升级菜单[%s]组件路径失败: %v", name, err)
		return
	}
	utils.Logger.Infof("已升级菜单[%s]组件路径: %s -> %s", name, legacyComponent, newComponent)
}

// upgradeMenuPath 幂等升级内置菜单的路由路径（口径同上）。
func upgradeMenuPath(name, legacyPath, newPath string) {
	var menu Menu
	if err := utils.DB.Where("name = ? AND path = ?", name, legacyPath).First(&menu).Error; err != nil {
		return
	}
	if err := utils.DB.Model(&menu).Update("path", newPath).Error; err != nil {
		utils.Logger.Warnf("升级菜单[%s]路由路径失败: %v", name, err)
		return
	}
	utils.Logger.Infof("已升级菜单[%s]路由路径: %s -> %s", name, legacyPath, newPath)
}

// defaultRoleMenus 描述内置角色初始化时应获得的菜单（按菜单名；父级目录由 GetUserMenus 自动补全，无需重复列出）。
// "*" 表示全部启用菜单。角色 1（管理员）本身即拥有全部菜单，无需在此声明。
// 注：原「普通管理员（admin）」角色已下线移除，其 "*"（全部菜单）配置一并删除。
var defaultRoleMenus = map[string][]string{
	"content_reviewer": {"管理首页", "待审核", "图文管理"},
	"content_author":   {"管理首页", "图文管理"},
}

// SeedDefaultRolePermissions 为内置角色初始化菜单权限：
// 仅在角色当前 permissions 为空时写入，避免覆盖管理员后续在「角色管理-分配权限」中的自定义配置。
// 否则全新部署时角色 3/4 没有任何菜单权限，登录后侧边栏为空且除少数豁免接口外全部返回「没有授权」。
func SeedDefaultRolePermissions() {
	for code, menuNames := range defaultRoleMenus {
		var role Role
		if err := utils.DB.Where("code = ?", code).First(&role).Error; err != nil {
			continue
		}
		if strings.TrimSpace(role.Permissions) != "" {
			continue
		}

		var menus []Menu
		query := utils.DB.Where("status = ?", 1)
		if !(len(menuNames) == 1 && menuNames[0] == "*") {
			query = query.Where("name IN ?", menuNames)
		}
		if err := query.Find(&menus).Error; err != nil {
			utils.Logger.Warnf("初始化默认角色[%s]菜单权限失败: %v", code, err)
			continue
		}
		if len(menus) == 0 {
			continue
		}

		ids := make([]string, 0, len(menus))
		for _, menu := range menus {
			ids = append(ids, strconv.FormatUint(uint64(menu.ID), 10))
		}
		if err := utils.DB.Model(&Role{}).Where("id = ?", role.ID).
			Update("permissions", strings.Join(ids, ",")).Error; err != nil {
			utils.Logger.Warnf("初始化默认角色[%s]菜单权限失败: %v", code, err)
			continue
		}
		utils.Logger.Infof("已初始化默认角色[%s]菜单权限: %d 项", code, len(menus))
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
