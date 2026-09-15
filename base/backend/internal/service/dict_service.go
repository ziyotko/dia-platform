package service

import (
	"base/internal/models"
	"base/pkg/db"

	"gorm.io/gorm"
)

type DictService struct{}

type DictListQuery struct {
	TenantID uint64
	Code     string
	Name     string
	Status   int
	Page     int
	Size     int
}

func (s DictService) Create(d *models.Dict) error {
	return db.DB.Create(d).Error
}

func (s DictService) Update(d *models.Dict, tenantID uint64) error {
	db := db.DB.Model(d).Where("id = ?", d.ID)
	if tenantID > 0 {
		db = db.Where("tenant_id = ?", tenantID)
	}
	return db.Updates(map[string]interface{}{
		"code":        d.Code,
		"name":        d.Name,
		"description": d.Description,
		"status":      d.Status,
	}).Error
}

func (s DictService) Delete(id uint64, tenantID uint64) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var d models.Dict
		query := tx.Where("id = ?", id)
		if tenantID > 0 {
			query = query.Where("tenant_id = ?", tenantID)
		}
		if err := query.First(&d).Error; err != nil {
			return err
		}
		if err := tx.Where("dict_id = ?", id).Delete(&models.DictItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Dict{}, id).Error
	})
}

func (s DictService) GetByID(id uint64, tenantID uint64) (*models.Dict, error) {
	var d models.Dict
	query := db.DB.Where("id = ?", id)
	// 读取口径：本租户 + 平台内置（tenant_id = 0）均可读；写操作仍限本租户
	if tenantID > 0 {
		query = query.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}
	err := query.Preload("Items").First(&d).Error
	return &d, err
}

func (s DictService) List(q DictListQuery) ([]models.Dict, int64, error) {
	var list []models.Dict
	var total int64
	query := db.DB.Model(&models.Dict{})
	if q.TenantID > 0 {
		// 读取口径：本租户 + 平台内置（tenant_id = 0）
		query = query.Where("tenant_id = ? OR tenant_id = 0", q.TenantID)
	}
	if q.Code != "" {
		query = query.Where("code LIKE ?", "%"+q.Code+"%")
	}
	if q.Name != "" {
		query = query.Where("name LIKE ?", "%"+q.Name+"%")
	}
	if q.Status >= 0 {
		query = query.Where("status = ?", q.Status)
	}
	query.Count(&total)
	offset := (q.Page - 1) * q.Size
	err := query.Order("created_at DESC").Offset(offset).Limit(q.Size).Find(&list).Error
	return list, total, err
}

// GetByCode 按编码取字典（含启用字典项）。
// 读取口径：租户自定义字典优先，无则回退平台内置字典（tenant_id = 0）。
func (s DictService) GetByCode(code string, tenantID uint64) (*models.Dict, error) {
	var list []models.Dict
	query := db.DB.Where("code = ? AND status = ?", code, 1)
	if tenantID > 0 {
		query = query.Where("tenant_id = ? OR tenant_id = 0", tenantID).Order("tenant_id DESC")
	}
	if err := query.Preload("Items", "status = ?", 1).Limit(1).Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &list[0], nil
}

func (s DictService) SaveItems(dictID uint64, items []models.DictItem, tenantID uint64) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		// 先校验字典归属，防止跨租户操作字典项
		var dict models.Dict
		dictQuery := tx.Where("id = ?", dictID)
		if tenantID > 0 {
			dictQuery = dictQuery.Where("tenant_id = ?", tenantID)
		}
		if err := dictQuery.First(&dict).Error; err != nil {
			return err
		}
		if err := tx.Where("dict_id = ?", dictID).Delete(&models.DictItem{}).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		for i := range items {
			items[i].DictID = dictID
			items[i].ID = 0
		}
		return tx.Create(&items).Error
	})
}
