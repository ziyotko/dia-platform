package service

import (
	"strconv"
	"time"

	"base/internal/models"
	"base/pkg/db"
)

type LoginLogService struct{}

type LoginLogListQuery struct {
	TenantID uint64
	Username string
	Status   int
	StartAt  string
	EndAt    string
	Page     int
	Size     int
}

func (s LoginLogService) Create(log *models.LoginLog) error {
	return db.DB.Create(log).Error
}

func (s LoginLogService) List(q LoginLogListQuery) ([]models.LoginLog, int64, error) {
	var list []models.LoginLog
	var total int64

	query := db.DB.Model(&models.LoginLog{})
	if q.TenantID > 0 {
		query = query.Where("tenant_id = ?", q.TenantID)
	}
	if q.Username != "" {
		query = query.Where("username LIKE ?", "%"+q.Username+"%")
	}
	if q.Status >= 0 {
		query = query.Where("status = ?", q.Status)
	}
	if q.StartAt != "" {
		query = query.Where("created_at >= ?", q.StartAt)
	}
	if q.EndAt != "" {
		query = query.Where("created_at <= ?", q.EndAt)
	}

	query.Count(&total)
	offset := (q.Page - 1) * q.Size
	err := query.Order("created_at DESC").Offset(offset).Limit(q.Size).Find(&list).Error
	return list, total, err
}

func (s LoginLogService) DeleteByIDs(ids []uint64, tenantID uint64) error {
	if len(ids) == 0 {
		return nil
	}
	db := db.DB.Where("id IN ?", ids)
	if tenantID > 0 {
		db = db.Where("tenant_id = ?", tenantID)
	}
	return db.Delete(&models.LoginLog{}).Error
}

func (s LoginLogService) ClearBefore(days int, tenantID uint64) error {
	deadline := time.Now().AddDate(0, 0, -days)
	db := db.DB.Where("created_at < ?", deadline)
	if tenantID > 0 {
		db = db.Where("tenant_id = ?", tenantID)
	}
	return db.Delete(&models.LoginLog{}).Error
}

func (s LoginLogService) Export(q LoginLogListQuery) (string, error) {
	q.Page = 1
	q.Size = 10000
	list, _, err := s.List(q)
	if err != nil {
		return "", err
	}

	var sb = newCSVBuilder()
	sb.WriteString("ID,租户ID,用户ID,用户名,IP,UserAgent,状态,消息,登录时间\n")
	for _, log := range list {
		status := "成功"
		if log.Status == 0 {
			status = "失败"
		}
		sb.WriteString(csvLine(
			strconv.FormatUint(log.ID, 10),
			strconv.FormatUint(log.TenantID, 10),
			strconv.FormatUint(log.UserID, 10),
			log.Username,
			log.IP,
			log.Agent,
			status,
			log.Message,
			log.CreatedAt.Format(time.DateTime),
		))
	}
	return sb.String(), nil
}
