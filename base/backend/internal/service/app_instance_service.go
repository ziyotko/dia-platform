package service

import (
	"errors"

	"base/internal/models"
	"base/pkg/db"

	"gorm.io/gorm/clause"
)

type AppInstanceService struct{}

// Create 开通应用。校验应用存在，且同一租户同一应用不能重复开通
// （删除是物理删除，删掉的历史实例不再占用该组合，因此删除后可以重新开通）。
func (s AppInstanceService) Create(i *models.AppInstance) error {
	if i.AppID == 0 {
		return errors.New("请选择应用")
	}
	var app models.App
	if err := db.DB.First(&app, i.AppID).Error; err != nil {
		return errors.New("应用不存在，请先在「应用管理」中创建")
	}
	var count int64
	if err := db.DB.Model(&models.AppInstance{}).
		Where("tenant_id = ? AND app_id = ?", i.TenantID, i.AppID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该租户已开通此应用，请勿重复开通")
	}
	// 必须 Omit 关联：AppInstance.App 是 belongs-to，GORM 的 Create 会用请求体里的 app 覆盖 base_app 行
	// （例如把应用的后端地址改成攻击者自己的服务），必须只写实例本身。
	return db.DB.Omit(clause.Associations).Create(i).Error
}

// IsEnabled 判断某租户是否已开通并启用了某应用（供子应用代理入口鉴权使用）。
func (s AppInstanceService) IsEnabled(tenantID, appID uint64) (bool, error) {
	var count int64
	err := db.DB.Model(&models.AppInstance{}).
		Where("tenant_id = ? AND app_id = ? AND status = ?", tenantID, appID, 1).
		Count(&count).Error
	return count > 0, err
}

func (s AppInstanceService) Update(i *models.AppInstance, tenantID uint64) error {
	check := db.DB.Model(&models.AppInstance{}).Where("id = ?", i.ID)
	db := db.DB.Model(i)
	if tenantID > 0 {
		check = check.Where("tenant_id = ?", tenantID)
		db = db.Where("tenant_id = ?", tenantID)
	}
	if err := ensureRecordExists(check, "应用实例不存在或不属于当前租户"); err != nil {
		return err
	}
	return db.Updates(map[string]interface{}{
		"status": i.Status,
		"config": i.Config,
	}).Error
}

func (s AppInstanceService) Delete(id uint64, tenantID uint64) error {
	query := db.DB.Where("id = ?", id)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	return ensureDeleteAffected(query.Delete(&models.AppInstance{}), "应用实例不存在或不属于当前租户")
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
