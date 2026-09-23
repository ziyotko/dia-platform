package service

import (
	"application/internal/models"
	"application/pkg/db"
)

type AuditService struct{}

func (s *AuditService) Record(adminID uint64, admin, module, action, ip, detail string) {
	db.DB.Create(&models.AuditLog{
		AdminID: adminID,
		Admin:   admin,
		Module:  module,
		Action:  action,
		IP:      ip,
		Detail:  detail,
	})
}

func (s *AuditService) List(page, size int, module, keyword string) ([]models.AuditLog, int64, error) {
	var list []models.AuditLog
	var total int64
	query := db.DB.Model(&models.AuditLog{})
	if module != "" {
		query = query.Where("module = ?", module)
	}
	if keyword != "" {
		query = query.Where("action LIKE ? OR admin LIKE ? OR detail LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

// maxAuditExportRows 导出上限：避免一次把整张表读进内存（同一筛选下的前 N 条）。
const maxAuditExportRows = 10000

// Export returns the rows matching the same filters the list uses, newest first.
func (s *AuditService) Export(module, keyword string) ([]models.AuditLog, error) {
	var list []models.AuditLog
	query := db.DB.Model(&models.AuditLog{})
	if module != "" {
		query = query.Where("module = ?", module)
	}
	if keyword != "" {
		query = query.Where("action LIKE ? OR admin LIKE ? OR detail LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	err := query.Order("created_at DESC").Limit(maxAuditExportRows).Find(&list).Error
	return list, err
}
