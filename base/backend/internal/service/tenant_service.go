package service

import (
	"errors"
	"fmt"

	"base/internal/models"
	"base/pkg/db"
	"base/pkg/utils"

	"gorm.io/gorm"
)

type TenantService struct{}

// Create 创建租户。
// 若存在同编码的软删除记录，直接恢复并覆盖字段：唯一索引在软删除后仍占位，
// 不这样做用户会看到数据库的 Duplicate entry 错误，且永远无法重建同编码租户。
func (s TenantService) Create(t *models.Tenant) error {
	if t.Code == "" {
		t.Code = "T" + utils.RandomDigit(8)
	}

	var existing models.Tenant
	err := db.DB.Unscoped().Where("code = ?", t.Code).First(&existing).Error
	switch {
	case err == nil:
		if !existing.DeletedAt.Valid {
			return errors.New("租户编码已存在")
		}
		if err := db.DB.Unscoped().Model(&models.Tenant{}).Where("id = ?", existing.ID).Updates(map[string]interface{}{
			"deleted_at":    nil,
			"name":          t.Name,
			"status":        t.Status,
			"contact_name":  t.ContactName,
			"contact_phone": t.ContactPhone,
			"description":   t.Description,
		}).Error; err != nil {
			return err
		}
		return db.DB.Where("id = ?", existing.ID).First(t).Error
	case errors.Is(err, gorm.ErrRecordNotFound):
		return db.DB.Create(t).Error
	default:
		return err
	}
}

func (s TenantService) Update(t *models.Tenant) error {
	if err := ensureRecordExists(db.DB.Model(&models.Tenant{}).Where("id = ?", t.ID), "租户不存在"); err != nil {
		return err
	}
	return db.DB.Model(t).Updates(map[string]interface{}{
		"name":          t.Name,
		"status":        t.Status,
		"contact_name":  t.ContactName,
		"contact_phone": t.ContactPhone,
		"description":   t.Description,
	}).Error
}

// Delete 删除租户。租户下还有业务数据时拒绝删除（避免产生孤儿数据），
// 需要先清理用户/角色/应用实例/流程数据。
func (s TenantService) Delete(id uint64) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Tenant{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errors.New("租户不存在")
		}

		refs := []struct {
			model interface{}
			name  string
		}{
			{&models.User{}, "用户"},
			{&models.Role{}, "角色"},
			{&models.AppInstance{}, "应用实例"},
			{&models.WorkflowRole{}, "流程角色"},
			{&models.Workflow{}, "流程定义"},
			{&models.WorkflowInstance{}, "流程实例"},
		}
		for _, ref := range refs {
			var n int64
			if err := tx.Model(ref.model).Where("tenant_id = ?", id).Count(&n).Error; err != nil {
				return err
			}
			if n > 0 {
				return fmt.Errorf("该租户下还有 %d 条%s数据，请先清理后再删除租户", n, ref.name)
			}
		}
		return tx.Where("id = ?", id).Delete(&models.Tenant{}).Error
	})
}

func (s TenantService) GetByID(id uint64) (*models.Tenant, error) {
	var t models.Tenant
	err := db.DB.First(&t, id).Error
	return &t, err
}

func (s TenantService) List(page, size int, keyword string) ([]models.Tenant, int64, error) {
	var list []models.Tenant
	var total int64
	query := db.DB.Model(&models.Tenant{})
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}
