package service

import (
	"conference/internal/models"
	"conference/pkg/db"
)

type AuditService struct{}

// Log records an audit entry
func (s *AuditService) Log(userID uint64, username, userType, action, resource string, resourceID uint64, detail, ip, userAgent string) {
	log := models.AuditLog{
		UserID:     userID,
		Username:   username,
		UserType:   userType,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Detail:     detail,
		IP:         ip,
		UserAgent:  userAgent,
	}
	db.DB.Create(&log)
}

// List returns paginated audit logs
func (s *AuditService) List(userType, action, resource string, page, size int) ([]models.AuditLog, int64, error) {
	var list []models.AuditLog
	var total int64
	query := db.DB.Model(&models.AuditLog{})
	if userType != "" {
		query = query.Where("user_type = ?", userType)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}
