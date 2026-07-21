package service

import (
	"base/internal/models"
	"base/pkg/db"
)

type AppInstanceService struct{}

func (s AppInstanceService) Create(i *models.AppInstance) error {
	return db.DB.Create(i).Error
}

func (s AppInstanceService) Update(i *models.AppInstance) error {
	return db.DB.Model(i).Updates(map[string]interface{}{
		"status": i.Status,
		"config": i.Config,
	}).Error
}

func (s AppInstanceService) Delete(id uint64) error {
	return db.DB.Delete(&models.AppInstance{BaseModel: models.BaseModel{ID: id}}).Error
}

func (s AppInstanceService) GetByID(id uint64) (*models.AppInstance, error) {
	var i models.AppInstance
	err := db.DB.Preload("App").First(&i, id).Error
	return &i, err
}

func (s AppInstanceService) ListByTenant(tenantID uint64, page, size int) ([]models.AppInstance, int64, error) {
	var list []models.AppInstance
	var total int64
	query := db.DB.Model(&models.AppInstance{}).Where("tenant_id = ?", tenantID).Preload("App")
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (s AppInstanceService) GetTenantActiveApps(tenantID uint64) ([]models.App, error) {
	var apps []models.App
	err := db.DB.Model(&models.App{}).
		Joins("JOIN base_app_instance ON base_app_instance.app_id = base_app.id").
		Where("base_app_instance.tenant_id = ? AND base_app_instance.status = ? AND base_app.status = ?", tenantID, 1, 1).
		Order("base_app.sort ASC").
		Find(&apps).Error
	return apps, err
}
