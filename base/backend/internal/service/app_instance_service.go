package service

import (
	"base/internal/models"
	"base/pkg/db"
)

type AppInstanceService struct{}

func (s AppInstanceService) Create(i *models.AppInstance) error {
	return db.DB.Create(i).Error
}

func (s AppInstanceService) Update(i *models.AppInstance, tenantID uint64) error {
	db := db.DB.Model(i)
	if tenantID > 0 {
		db = db.Where("tenant_id = ?", tenantID)
	}
	return db.Updates(map[string]interface{}{
		"status": i.Status,
		"config": i.Config,
	}).Error
}

func (s AppInstanceService) Delete(id uint64, tenantID uint64) error {
	db := db.DB
	if tenantID > 0 {
		db = db.Where("tenant_id = ?", tenantID)
	}
	return db.Delete(&models.AppInstance{BaseModel: models.BaseModel{ID: id}}).Error
}

func (s AppInstanceService) GetByID(id uint64, tenantID uint64) (*models.AppInstance, error) {
	var i models.AppInstance
	query := db.DB.Preload("App").Where("id = ?", id)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	err := query.First(&i).Error
	return &i, err
}

// List 分页查询应用实例。tenantID 为当前登录用户所属租户：
//   - 普通租户用户（tenantID > 0）只能看到本租户；
//   - 平台超管（tenantID == 0）不传 filterTenantID 时查看全部租户，传则只看指定租户。
func (s AppInstanceService) List(tenantID, filterTenantID uint64, page, size int) ([]models.AppInstance, int64, error) {
	var list []models.AppInstance
	var total int64
	query := db.DB.Model(&models.AppInstance{}).Preload("App")
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	} else if filterTenantID > 0 {
		query = query.Where("tenant_id = ?", filterTenantID)
	}
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
