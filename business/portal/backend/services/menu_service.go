package services

import (
	"errors"
	"slices"
	"sort"
	"strconv"
	"strings"

	"server/models"
	"server/utils"
)

type MenuService struct{}

func (s *MenuService) GetAllMenus() ([]models.Menu, error) {
	var menus []models.Menu
	err := utils.DB.Order("sort ASC, id ASC").Find(&menus).Error
	if err != nil {
		return nil, err
	}
	return buildMenuTree(menus, 0), nil
}

func (s *MenuService) GetUserMenus(userID uint) ([]models.Menu, error) {
	userService := &UserService{}
	roleIds, err := userService.GetUserRoleIds(userID)
	if err != nil {
		return nil, err
	}

	permMap := make(map[uint]bool)
	roleService := &RoleService{}
	for _, roleId := range roleIds {
		perms, err := roleService.GetRolePermissions(uint(roleId))
		if err != nil {
			continue
		}
		for _, pid := range perms {
			permMap[pid] = true
		}
	}

	var allMenus []models.Menu
	err = utils.DB.Where("status = ?", 1).Order("sort ASC, id ASC").Find(&allMenus).Error
	if err != nil {
		return nil, err
	}

	// 管理员（角色 ID=1）默认拥有一切菜单权限，直接返回全部启用菜单；
	// 其余角色的菜单由「角色管理-分配权限」决定，以便按机构/职责范围收缩可见范围。
	if slices.Contains(roleIds, models.RoleIDSuperAdmin) {
		return buildMenuTree(allMenus, 0), nil
	}

	menuMap := make(map[uint]models.Menu)
	for _, m := range allMenus {
		menuMap[m.ID] = m
	}

	keepMap := make(map[uint]bool)
	for _, m := range allMenus {
		if permMap[m.ID] {
			cur := m
			for {
				keepMap[cur.ID] = true
				if cur.ParentID == 0 {
					break
				}
				parent, ok := menuMap[cur.ParentID]
				if !ok {
					break
				}
				cur = parent
			}
		}
	}

	var filtered []models.Menu
	for _, m := range allMenus {
		if keepMap[m.ID] {
			filtered = append(filtered, m)
		}
	}

	return buildMenuTree(filtered, 0), nil
}

func (s *MenuService) CreateMenu(menu *models.Menu) error {
	// 上级菜单必须存在（否则该菜单成为孤児，只能靠 buildMenuTree 提升为根，与 Update 侧口径不一致）
	if err := validateTreeParent("menu", 0, menu.ParentID); err != nil {
		return err
	}
	return utils.DB.Create(menu).Error
}

func (s *MenuService) UpdateMenu(id uint, menu *models.Menu) error {
	if err := validateTreeParent("menu", id, menu.ParentID); err != nil {
		return err
	}
	return utils.DB.Model(&models.Menu{}).Where("id = ?", id).Updates(map[string]any{
		"parent_id":  menu.ParentID,
		"name":       menu.Name,
		"path":       menu.Path,
		"component":  menu.Component,
		"api_prefix": menu.APIPrefix,
		"icon":       menu.Icon,
		"type":       menu.Type,
		"sort":       menu.Sort,
		"status":     menu.Status,
	}).Error
}

func (s *MenuService) DeleteMenu(id uint) error {
	var count int64
	utils.DB.Model(&models.Menu{}).Where("parent_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("该菜单下存在子菜单，无法删除")
	}
	if err := utils.DB.Delete(&models.Menu{}, id).Error; err != nil {
		return err
	}
	// 菜单被删除后，role.permissions 中会残留该 ID：\n	// 它在权限树里已无对应节点，既无法再次勾选也无法取消，只能靠重新保存角色清理。
	// 这里在删除后同步剔除各角色对该菜单的引用（保留顺序、去重）。
	var roles []models.Role
	if err := utils.DB.Find(&roles).Error; err != nil {
		return nil // 菜单已删除，权限清理失败不影响主流程
	}
	removed := strconv.FormatUint(uint64(id), 10)
	for i := range roles {
		role := roles[i]
		if strings.TrimSpace(role.Permissions) == "" {
			continue
		}
		parts := strings.Split(role.Permissions, ",")
		kept := make([]string, 0, len(parts))
		changed := false
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed == "" {
				continue
			}
			if trimmed == removed {
				changed = true
				continue
			}
			kept = append(kept, trimmed)
		}
		if !changed {
			continue
		}
		if err := utils.DB.Model(&models.Role{}).Where("id = ?", role.ID).
			Update("permissions", strings.Join(kept, ",")).Error; err != nil {
			utils.Logger.Warnf("清理角色[%s]菜单权限失败: %v", role.Code, err)
		}
	}
	return nil
}

// buildMenuTree 将扁平菜单列表组装为层级树。
// 使用「按父节点分组 + 访问标记」迭代实现：既避免重复遍历，也能防御脏数据成环导致的无限递归；
// 父节点缺失的“孤儿”节点会被提升为根节点，避免从树中静默消失。
func buildMenuTree(menus []models.Menu, parentID uint) []models.Menu {
	childrenByParent := make(map[uint][]models.Menu, len(menus))
	for _, menu := range menus {
		childrenByParent[menu.ParentID] = append(childrenByParent[menu.ParentID], menu)
	}

	visited := make(map[uint]bool, len(menus))
	var build func(pid uint) []models.Menu
	build = func(pid uint) []models.Menu {
		var tree []models.Menu
		for _, menu := range childrenByParent[pid] {
			if visited[menu.ID] {
				continue // 脏数据成环，跳过已访问节点，防止无限递归
			}
			visited[menu.ID] = true
			menu.Children = build(menu.ID)
			tree = append(tree, menu)
		}
		sortMenusBySort(tree)
		return tree
	}

	tree := build(parentID)

	// 兜底：父节点不存在的孤儿节点提升为根，保证不因数据异常而整体丢失
	for _, menu := range menus {
		if menu.ParentID == parentID || visited[menu.ID] {
			continue
		}
		visited[menu.ID] = true
		menu.Children = build(menu.ID)
		tree = append(tree, menu)
	}
	sortMenusBySort(tree)
	return tree
}

func sortMenusBySort(menus []models.Menu) {
	sort.Slice(menus, func(i, j int) bool {
		return menus[i].Sort < menus[j].Sort
	})
}
