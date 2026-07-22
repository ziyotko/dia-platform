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

func (s RoleService) GetByID(id uint64) (*models.Role, error) {
	var r models.Role
	err := db.DB.Preload("Menus").Preload("Perms").First(&r, id).Error
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

func (s RoleService) AssignMenus(roleID uint64, menuIDs []uint64) error {
	var role models.Role
	if err := db.DB.First(&role, roleID).Error; err != nil {
		return err
	}
	var menus []models.Menu
	if len(menuIDs) > 0 {
		if err := db.DB.Find(&menus, menuIDs).Error; err != nil {
			return err
		}
	}
	return db.DB.Model(&role).Association("Menus").Replace(menus)
}

func (s RoleService) AssignPermissions(roleID uint64, permIDs []uint64) error {
	var role models.Role
	if err := db.DB.First(&role, roleID).Error; err != nil {
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
