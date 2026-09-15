package service

import (
	"base/internal/models"
	"base/pkg/db"
)

type DashboardService struct{}

type StatsResult struct {
	TenantCount       int64 `json:"tenantCount"`
	AppCount          int64 `json:"appCount"`
	UserCount         int64 `json:"userCount"`
	RoleCount         int64 `json:"roleCount"`
	OrganizationCount int64 `json:"organizationCount"`
	MessageCount      int64 `json:"messageCount"`
}

// GetStats 仪表盘统计。
// 口径：平台超管看全平台；租户用户只看本租户（含平台级 tenant_id = 0 的共享数据），
// 租户数/应用数对租户无意义，返回 0（前端对非超管隐藏这两张卡片）。
func (s DashboardService) GetStats(tenantID uint64) (*StatsResult, error) {
	var res StatsResult

	if models.IsPlatformTenant(tenantID) {
		if err := db.DB.Model(&models.Tenant{}).Count(&res.TenantCount).Error; err != nil {
			return nil, err
		}
		if err := db.DB.Model(&models.App{}).Count(&res.AppCount).Error; err != nil {
			return nil, err
		}
		if err := db.DB.Model(&models.User{}).Count(&res.UserCount).Error; err != nil {
			return nil, err
		}
		if err := db.DB.Model(&models.Role{}).Count(&res.RoleCount).Error; err != nil {
			return nil, err
		}
		if err := db.DB.Model(&models.Organization{}).Count(&res.OrganizationCount).Error; err != nil {
			return nil, err
		}
		if err := db.DB.Model(&models.Message{}).Count(&res.MessageCount).Error; err != nil {
			return nil, err
		}
		return &res, nil
	}

	// 租户用户：已开通且启用的应用数
	if err := db.DB.Model(&models.App{}).
		Joins("JOIN base_app_instance ON base_app_instance.app_id = base_app.id").
		Where("base_app_instance.tenant_id = ? AND base_app_instance.status = ? AND base_app.status = ?", tenantID, 1, 1).
		Count(&res.AppCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Model(&models.User{}).Where("tenant_id = ?", tenantID).Count(&res.UserCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Model(&models.Role{}).Where("tenant_id = ?", tenantID).Count(&res.RoleCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Model(&models.Organization{}).Where("tenant_id = ? OR tenant_id = 0", tenantID).Count(&res.OrganizationCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Model(&models.Message{}).Where("tenant_id = ? OR tenant_id = 0", tenantID).Count(&res.MessageCount).Error; err != nil {
		return nil, err
	}
	return &res, nil
}
