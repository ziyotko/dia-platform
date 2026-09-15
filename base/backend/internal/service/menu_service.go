package service

import (
	"base/internal/models"
	"base/pkg/db"
)

type MenuService struct{}

func (s MenuService) Create(m *models.Menu) error {
	return db.DB.Create(m).Error
}

func (s MenuService) Update(m *models.Menu, tenantID uint64) error {
	db := db.DB.Model(m)
	if tenantID > 0 {
		db = db.Where("tenant_id = ?", tenantID)
	}
	return db.Updates(map[string]interface{}{
		"parent_id":  m.ParentID,
		"app_code":   m.AppCode,
		"name":       m.Name,
		"icon":       m.Icon,
		"path":       m.Path,
		"component":  m.Component,
		"type":       m.Type,
		"permission": m.Permission,
		"sort":       m.Sort,
		"status":     m.Status,
		"hidden":     m.Hidden,
		"keep_alive": m.KeepAlive,
		"target":     m.Target,
	}).Error
}

func (s MenuService) Delete(id uint64, tenantID uint64) error {
	db := db.DB
	if tenantID > 0 {
		db = db.Where("tenant_id = ?", tenantID)
	}
	return db.Delete(&models.Menu{BaseModel: models.BaseModel{ID: id}}).Error
}

func (s MenuService) GetByID(id uint64, tenantID uint64) (*models.Menu, error) {
	var m models.Menu
	query := db.DB.Where("id = ?", id)
	// 读取口径：本租户 + 平台内置（tenant_id = 0）均可读；写操作仍限本租户
	if tenantID > 0 {
		query = query.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}
	err := query.First(&m).Error
	return &m, err
}

func (s MenuService) List(tenantID uint64, appCode string) ([]models.Menu, error) {
	var list []models.Menu
	query := db.DB.Model(&models.Menu{}).Where("tenant_id = ? OR tenant_id = 0", tenantID)
	if appCode != "" {
		query = query.Where("app_code = ?", appCode)
	}
	err := query.Order("sort ASC, created_at ASC").Find(&list).Error
	return list, err
}

func (s MenuService) GetTree(tenantID uint64, appCode string) ([]models.Menu, error) {
	list, err := s.List(tenantID, appCode)
	if err != nil {
		return nil, err
	}
	return buildMenuTree(list, 0), nil
}

func buildMenuTree(list []models.Menu, parentID uint64) []models.Menu {
	tree := make([]models.Menu, 0)
	for _, m := range list {
		if m.ParentID == parentID {
			children := buildMenuTree(list, m.ID)
			if len(children) > 0 {
				m.Children = children
			}
			tree = append(tree, m)
		}
	}
	return tree
}

// GetUserMenus 按用户角色聚合可见菜单树。
// 平台超管（tenantID == 0）与租户管理员（is_admin）返回租户可见的全部启用菜单。
func (s MenuService) GetUserMenus(userID, tenantID uint64) ([]models.Menu, error) {
	// 通过用户的角色聚合菜单
	var user models.User
	if err := db.DB.Preload("Roles.Menus").First(&user, userID).Error; err != nil {
		return nil, err
	}
	// 超级管理员或租户管理员返回全部启用菜单
	if user.IsAdmin {
		all, err := s.List(tenantID, "")
		if err != nil {
			return nil, err
		}
		enabled := make([]models.Menu, 0, len(all))
		for _, m := range all {
			if m.Status == 1 {
				enabled = append(enabled, m)
			}
		}
		return buildMenuTree(enabled, 0), nil
	}

	menuMap := make(map[uint64]models.Menu)
	for _, role := range user.Roles {
		for _, m := range role.Menus {
			if m.Status == 1 && (m.TenantID == 0 || m.TenantID == tenantID) {
				menuMap[m.ID] = m
			}
		}
	}
	list := make([]models.Menu, 0, len(menuMap))
	for _, m := range menuMap {
		list = append(list, m)
	}

	// 容错：历史数据可能只给角色分配了子菜单，需补齐父级，否则整棵子树不会出现在菜单树里
	list, err := completeMenuAncestors(list)
	if err != nil {
		return nil, err
	}
	visible := make([]models.Menu, 0, len(list))
	for _, m := range list {
		if m.TenantID == 0 || m.TenantID == tenantID {
			visible = append(visible, m)
		}
	}
	return buildMenuTree(visible, 0), nil
}

// completeMenuAncestors 补齐给定菜单的所有祖先节点（按 id 去重）。
func completeMenuAncestors(menus []models.Menu) ([]models.Menu, error) {
	var all []models.Menu
	if err := db.DB.Find(&all).Error; err != nil {
		return nil, err
	}
	byID := make(map[uint64]models.Menu, len(all))
	for _, m := range all {
		byID[m.ID] = m
	}

	result := make(map[uint64]models.Menu, len(menus))
	for _, m := range menus {
		result[m.ID] = m
	}
	for _, m := range menus {
		parentID := m.ParentID
		for parentID != 0 {
			if _, ok := result[parentID]; ok {
				break
			}
			parent, ok := byID[parentID]
			if !ok {
				break
			}
			result[parentID] = parent
			parentID = parent.ParentID
		}
	}

	out := make([]models.Menu, 0, len(result))
	for _, m := range result {
		out = append(out, m)
	}
	return out, nil
}
