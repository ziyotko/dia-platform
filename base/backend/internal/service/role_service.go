package service

import (
	"errors"

	"base/internal/models"
	"base/pkg/db"
)

type RoleService struct{}

func (s RoleService) Create(r *models.Role) error {
	return db.DB.Create(r).Error
}

func (s RoleService) Update(r *models.Role, tenantID uint64) error {
	check := db.DB.Model(&models.Role{}).Where("id = ?", r.ID)
	db := db.DB.Model(r)
	if tenantID > 0 {
		check = check.Where("tenant_id = ?", tenantID)
		db = db.Where("tenant_id = ?", tenantID)
	}
	if err := ensureRecordExists(check, "角色不存在或不属于当前租户"); err != nil {
		return err
	}
	return db.Updates(map[string]interface{}{
		"name":   r.Name,
		"code":   r.Code,
		"status": r.Status,
		"remark": r.Remark,
	}).Error
}

func (s RoleService) Delete(id uint64, tenantID uint64) error {
	query := db.DB.Where("id = ?", id)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	var role models.Role
	if err := query.First(&role).Error; err != nil {
		return err
	}
	// 平台内置超级管理员角色是权限基线，不允许删除
	if role.TenantID == 0 && role.Code == "super_admin" {
		return errors.New("平台超级管理员角色不允许删除")
	}
	return db.DB.Delete(&role).Error
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

// List 分页查询角色。tenantID 为当前登录用户所属租户：
//   - 普通租户用户（tenantID > 0）只能看到本租户；
//   - 平台超管（tenantID == 0）不传 filterTenantID 时查看全部租户，传则只看指定租户。
func (s RoleService) List(tenantID, filterTenantID uint64, page, size int, keyword string) ([]models.Role, int64, error) {
	var list []models.Role
	var total int64
	query := db.DB.Model(&models.Role{})
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	} else if filterTenantID > 0 {
		query = query.Where("tenant_id = ?", filterTenantID)
	}
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
		// 菜单树只会上报全选节点，目录类父节点往往是半选状态，需要补齐祖先节点
		completed, err := completeMenuAncestors(menus)
		if err != nil {
			return err
		}
		menus = completed
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
