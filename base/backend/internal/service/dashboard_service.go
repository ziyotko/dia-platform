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

func (s DashboardService) GetStats(tenantID uint64) (*StatsResult, error) {
	var res StatsResult
	if err := db.DB.Model(&models.Tenant{}).Count(&res.TenantCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Model(&models.App{}).Count(&res.AppCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Model(&models.User{}).Where("tenant_id = ? OR ? = 0", tenantID, tenantID).Count(&res.UserCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Model(&models.Role{}).Where("tenant_id = ? OR ? = 0", tenantID, tenantID).Count(&res.RoleCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Model(&models.Organization{}).Where("tenant_id = ? OR ? = 0", tenantID, tenantID).Count(&res.OrganizationCount).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Model(&models.Message{}).Where("tenant_id = ? OR ? = 0", tenantID, tenantID).Count(&res.MessageCount).Error; err != nil {
		return nil, err
	}
	return &res, nil
}
