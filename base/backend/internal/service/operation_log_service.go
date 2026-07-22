package service

import (
	"fmt"
	"strings"
	"time"

	"base/internal/models"
	"base/pkg/db"
)

type OperationLogService struct{}

type LogListQuery struct {
	TenantID uint64
	UserID   uint64
	Username string
	Module   string
	Action   string
	Method   string
	Path     string
	Status   int
	StartAt  string
	EndAt    string
	Page     int
	Size     int
}

func (s OperationLogService) List(q LogListQuery) ([]models.OperationLog, int64, error) {
	var list []models.OperationLog
	var total int64

	query := db.DB.Model(&models.OperationLog{})
	if q.TenantID > 0 {
		query = query.Where("tenant_id = ?", q.TenantID)
	}
	if q.UserID > 0 {
		query = query.Where("user_id = ?", q.UserID)
	}
	if q.Username != "" {
		query = query.Where("username LIKE ?", "%"+q.Username+"%")
	}
	if q.Module != "" {
		query = query.Where("module LIKE ?", "%"+q.Module+"%")
	}
	if q.Action != "" {
		query = query.Where("action LIKE ?", "%"+q.Action+"%")
	}
	if q.Method != "" {
		query = query.Where("method = ?", q.Method)
	}
	if q.Path != "" {
		query = query.Where("path LIKE ?", "%"+q.Path+"%")
	}
	if q.Status >= 0 {
		query = query.Where("status = ?", q.Status)
	}
	if q.StartAt != "" {
		query = query.Where("operation_at >= ?", q.StartAt)
	}
	if q.EndAt != "" {
		query = query.Where("operation_at <= ?", q.EndAt)
	}

	query.Count(&total)
	offset := (q.Page - 1) * q.Size
	err := query.Order("operation_at DESC").Offset(offset).Limit(q.Size).Find(&list).Error
	return list, total, err
}

func (s OperationLogService) DeleteByIDs(ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	return db.DB.Where("id IN ?", ids).Delete(&models.OperationLog{}).Error
}

func (s OperationLogService) ClearBefore(days int, tenantID uint64) error {
	deadline := time.Now().AddDate(0, 0, -days)
	db := db.DB.Where("operation_at < ?", deadline)
	if tenantID > 0 {
		db = db.Where("tenant_id = ?", tenantID)
	}
	return db.Delete(&models.OperationLog{}).Error
}

func (s OperationLogService) Export(q LogListQuery) (string, error) {
	q.Page = 1
	q.Size = 10000
	list, _, err := s.List(q)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("ID,租户ID,用户ID,用户名,模块,操作,方法,路径,IP,状态,耗时(ms),操作时间\n")
	for _, log := range list {
		status := "成功"
		if log.Status == 0 {
			status = "失败"
		}
		sb.WriteString(fmt.Sprintf("%d,%d,%d,%s,%s,%s,%s,%s,%s,%s,%d,%s\n",
			log.ID,
			log.TenantID,
			log.UserID,
			log.Username,
			log.Module,
			log.Action,
			log.Method,
			log.Path,
			log.IP,
			status,
			log.Duration,
			log.OperationAt.Format(time.DateTime),
		))
	}
	return sb.String(), nil
}
