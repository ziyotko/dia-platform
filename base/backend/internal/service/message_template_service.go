package service

import (
	"base/internal/models"
	"base/pkg/db"
)

type MessageTemplateService struct{}

func (s MessageTemplateService) Create(t *models.MessageTemplate) error {
	return db.DB.Create(t).Error
}

func (s MessageTemplateService) Update(t *models.MessageTemplate) error {
	return db.DB.Model(t).Updates(map[string]interface{}{
		"name":        t.Name,
		"channel":     t.Channel,
		"subject":     t.Subject,
		"content":     t.Content,
		"variables":   t.Variables,
		"status":      t.Status,
		"description": t.Description,
	}).Error
}

func (s MessageTemplateService) Delete(id uint64) error {
	return db.DB.Delete(&models.MessageTemplate{BaseModel: models.BaseModel{ID: id}}).Error
}

func (s MessageTemplateService) GetByID(id uint64) (*models.MessageTemplate, error) {
	var t models.MessageTemplate
	err := db.DB.First(&t, id).Error
	return &t, err
}

func (s MessageTemplateService) List(page, size int, keyword string) ([]models.MessageTemplate, int64, error) {
	var list []models.MessageTemplate
	var total int64
	query := db.DB.Model(&models.MessageTemplate{})
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}
