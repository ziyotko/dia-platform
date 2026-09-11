package service

import (
	"member/internal/models"
	"member/pkg/db"
)

type OperationLogService struct{}

// List returns paginated operation logs (admin).
func (s *OperationLogService) List(page, size int, keyword string) ([]models.OperationLog, int64, error) {
	var list []models.OperationLog
	var total int64

	query := db.DB.Model(&models.OperationLog{})
	if keyword != "" {
		kw := "%" + keyword + "%"
		query = query.Where("username LIKE ? OR path LIKE ? OR method LIKE ? OR ip LIKE ?",
			kw, kw, kw, kw)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
