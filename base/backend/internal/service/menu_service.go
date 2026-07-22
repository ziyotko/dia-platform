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
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
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
	var tree []models.Menu
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

func (s MenuService) GetUserMenus(userID, tenantID uint64) ([]models.Menu, error) {
	// 通过用户的角色聚合菜单
	var user models.User
	if err := db.DB.Preload("Roles.Menus").First(&user, userID).Error; err != nil {
		return nil, err
	}
	// 超级管理员或租户管理员返回所有可用菜单
	if user.IsAdmin {
		return s.GetTree(tenantID, "")
	}
	menuMap := make(map[uint64]models.Menu)
	for _, role := range user.Roles {
		for _, m := range role.Menus {
			if m.Status == 1 && (m.TenantID == 0 || m.TenantID == tenantID) {
				menuMap[m.ID] = m
			}
		}
	}
	var list []models.Menu
	for _, m := range menuMap {
		list = append(list, m)
	}
	return buildMenuTree(list, 0), nil
}
