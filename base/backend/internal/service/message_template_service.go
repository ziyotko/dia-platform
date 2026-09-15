package service

import (
	"base/internal/models"
	"base/pkg/db"
)

type MessageTemplateService struct{}

func (s MessageTemplateService) Create(t *models.MessageTemplate) error {
	return db.DB.Create(t).Error
}

func (s MessageTemplateService) Update(t *models.MessageTemplate, tenantID uint64) error {
	check := db.DB.Model(&models.MessageTemplate{}).Where("id = ?", t.ID)
	db := db.DB.Model(t).Where("id = ?", t.ID)
	if tenantID > 0 {
		check = check.Where("tenant_id = ?", tenantID)
		db = db.Where("tenant_id = ?", tenantID)
	}
	if err := ensureRecordExists(check, msgNotOwnedOrMissing); err != nil {
		return err
	}
	return db.Updates(map[string]interface{}{
		"name":        t.Name,
		"channel":     t.Channel,
		"subject":     t.Subject,
		"content":     t.Content,
		"variables":   t.Variables,
		"status":      t.Status,
		"description": t.Description,
	}).Error
}

func (s MessageTemplateService) Delete(id uint64, tenantID uint64) error {
	db := db.DB.Where("id = ?", id)
	if tenantID > 0 {
		db = db.Where("tenant_id = ?", tenantID)
	}
	return ensureDeleteAffected(db.Delete(&models.MessageTemplate{}), msgNotOwnedOrMissingDelete)
}

func (s MessageTemplateService) GetByID(id uint64, tenantID uint64) (*models.MessageTemplate, error) {
	var t models.MessageTemplate
	query := db.DB.Where("id = ?", id)
	// 读取口径：本租户 + 平台内置（tenant_id = 0）均可读；写操作仍限本租户
	if tenantID > 0 {
		query = query.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}
	err := query.First(&t).Error
	return &t, err
}

func (s MessageTemplateService) List(page, size int, keyword string, tenantID uint64) ([]models.MessageTemplate, int64, error) {
	var list []models.MessageTemplate
	var total int64
	query := db.DB.Model(&models.MessageTemplate{})
	if tenantID > 0 {
		// 读取口径：本租户 + 平台内置（tenant_id = 0）
		query = query.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}
