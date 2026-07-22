package service

import (
	"base/internal/models"
	"base/pkg/db"
)

type RoleService struct{}

func (s RoleService) Create(r *models.Role) error {
	return db.DB.Create(r).Error
}

func (s RoleService) Update(r *models.Role, tenantID uint64) error {
	db := db.DB.Model(r)
	if tenantID > 0 {
		db = db.Where("tenant_id = ?", tenantID)
	}
	return db.Updates(map[string]interface{}{
		"name":   r.Name,
		"code":   r.Code,
		"status": r.Status,
		"remark": r.Remark,
	}).Error
}

func (s RoleService) Delete(id uint64, tenantID uint64) error {
	db := db.DB
	if tenantID > 0 {
		db = db.Where("tenant_id = ?", tenantID)
	}
	return db.Delete(&models.Role{BaseModel: models.BaseModel{ID: id}}).Error
}

func (s RoleService) GetByID(id uint64, tenantID uint64) (*models.Role, error) {
	var r models.Role
	query := db.DB.Preload("Menus").Preload("Perms").Where("id = ?", id)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	err := query.First(&r).Error
	return &r, err
}

func (s RoleService) List(tenantID uint64, page, size int, keyword string) ([]models.Role, int64, error) {
	var list []models.Role
	var total int64
	query := db.DB.Model(&models.Role{}).Where("tenant_id = ?", tenantID)
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (s RoleService) AssignMenus(roleID uint64, menuIDs []uint64, tenantID uint64) error {
	var role models.Role
	query := db.DB.Where("id = ?", roleID)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&role).Error; err != nil {
		return err
	}
	var menus []models.Menu
	if len(menuIDs) > 0 {
		menuQuery := db.DB.Where("id IN ?", menuIDs)
		if tenantID > 0 {
			menuQuery = menuQuery.Where("tenant_id = ? OR tenant_id = 0", tenantID)
		}
		if err := menuQuery.Find(&menus).Error; err != nil {
			return err
		}
	}
	return db.DB.Model(&role).Association("Menus").Replace(menus)
}

func (s RoleService) AssignPermissions(roleID uint64, permIDs []uint64, tenantID uint64) error {
	var role models.Role
	query := db.DB.Where("id = ?", roleID)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&role).Error; err != nil {
		return err
	}
	var perms []models.Permission
	if len(permIDs) > 0 {
		if err := db.DB.Find(&perms, permIDs).Error; err != nil {
			return err
		}
	}
	return db.DB.Model(&role).Association("Perms").Replace(perms)
}
