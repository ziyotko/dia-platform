package services

import (
	"errors"
	"sort"

	"server/models"
	"server/utils"
)

type MenuService struct{}

func (s *MenuService) GetMenuList() ([]models.Menu, error) {
	var menus []models.Menu
	err := utils.DB.Where("status = ?", 1).Order("sort ASC, id ASC").Find(&menus).Error
	if err != nil {
		return nil, err
	}
	return buildMenuTree(menus, 0), nil
}

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

func (s *MenuService) GetMenuByID(id uint) (*models.Menu, error) {
	var menu models.Menu
	if err := utils.DB.First(&menu, id).Error; err != nil {
		return nil, errors.New("菜单不存在")
	}
	return &menu, nil
}

func (s *MenuService) CreateMenu(menu *models.Menu) error {
	return utils.DB.Create(menu).Error
}

func (s *MenuService) UpdateMenu(id uint, menu *models.Menu) error {
	return utils.DB.Model(&models.Menu{}).Where("id = ?", id).Updates(map[string]interface{}{
		"parent_id": menu.ParentID,
		"name":      menu.Name,
		"path":      menu.Path,
		"component": menu.Component,
		"icon":      menu.Icon,
		"type":      menu.Type,
		"sort":      menu.Sort,
		"status":    menu.Status,
	}).Error
}

func (s *MenuService) DeleteMenu(id uint) error {
	var count int64
	utils.DB.Model(&models.Menu{}).Where("parent_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("该菜单下存在子菜单，无法删除")
	}
	return utils.DB.Delete(&models.Menu{}, id).Error
}

func buildMenuTree(menus []models.Menu, parentID uint) []models.Menu {
	var tree []models.Menu
	for _, menu := range menus {
		if menu.ParentID == parentID {
			children := buildMenuTree(menus, menu.ID)
			menu.Children = children
			tree = append(tree, menu)
		}
	}
	sort.Slice(tree, func(i, j int) bool {
		return tree[i].Sort < tree[j].Sort
	})
	return tree
}
