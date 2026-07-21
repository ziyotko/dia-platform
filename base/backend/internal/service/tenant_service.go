package service

import (
	"base/internal/models"
	"base/pkg/db"
	"base/pkg/utils"
)

type TenantService struct{}

func (s TenantService) Create(t *models.Tenant) error {
	if t.Code == "" {
		t.Code = "T" + utils.RandomDigit(8)
	}
	return db.DB.Create(t).Error
}

func (s TenantService) Update(t *models.Tenant) error {
	return db.DB.Model(t).Updates(map[string]interface{}{
		"name":          t.Name,
		"status":        t.Status,
		"contact_name":  t.ContactName,
		"contact_phone": t.ContactPhone,
		"description":   t.Description,
	}).Error
}

func (s TenantService) Delete(id uint64) error {
	return db.DB.Delete(&models.Tenant{BaseModel: models.BaseModel{ID: id}}).Error
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
