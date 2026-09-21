package services

import (
	"time"

	"server/models"
	"server/utils"
)

type LogService struct{}

type LogListResult struct {
	Total int64                 `json:"total"`
	List  []models.OperationLog `json:"list"`
}

func (s *LogService) GetLogList(page, pageSize int, username, logType string, startDate, endDate string) (*LogListResult, error) {
	var logs []models.OperationLog
	var total int64

	query := utils.DB.Model(&models.OperationLog{})

	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if logType != "" {
		query = query.Where("type = ?", logType)
	}
	if startDate != "" && endDate != "" {
		query = query.Where("DATE(created_at) BETWEEN ? AND ?", startDate, endDate)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	err = query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	if err != nil {
		return nil, err
	}

	return &LogListResult{
		Total: total,
		List:  logs,
	}, nil
}

// ClearLogs 仅清空半年前的日志，系统必须保留最近半年的日志。返回删除条数。
func (s *LogService) ClearLogs() (int64, error) {
	cutoff := time.Now().AddDate(0, -6, 0)
	result := utils.DB.Where("created_at < ?", cutoff).Unscoped().Delete(&models.OperationLog{})
	return result.RowsAffected, result.Error
}

func (s *LogService) CreateLog(log *models.OperationLog) error {
	return utils.DB.Create(log).Error
}

func (s *LogService) GetUserOperationCount(userID uint) (int64, error) {
	var count int64
	err := utils.DB.Model(&models.OperationLog{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

type LoginLogListResult struct {
	Total int64             `json:"total"`
	List  []models.LoginLog `json:"list"`
}

func (s *LogService) GetLoginLogList(page, pageSize int, username, status, startDate, endDate string) (*LoginLogListResult, error) {
	var logs []models.LoginLog
	var total int64

	query := utils.DB.Model(&models.LoginLog{})

	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if startDate != "" && endDate != "" {
		query = query.Where("DATE(created_at) BETWEEN ? AND ?", startDate, endDate)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	err = query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	if err != nil {
		return nil, err
	}

	return &LoginLogListResult{
		Total: total,
		List:  logs,
	}, nil
}

// GetRecentLoginLogs 最近登录日志。
// usernames 非空时只返回这些用户名的记录（登录成功记录显示名、失败记录登录输入值，故调用方传多个候选）。
func (s *LogService) GetRecentLoginLogs(limit int, usernames []string) ([]models.LoginLog, error) {
	var logs []models.LoginLog
	query := utils.DB.Model(&models.LoginLog{})
	if len(usernames) > 0 {
		query = query.Where("username IN ?", usernames)
	}
	err := query.Order("id DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

// ClearLoginLogs 与 ClearLogs 保持同一语义：仅清空半年前的登录日志，保留最近半年。返回删除条数。
func (s *LogService) ClearLoginLogs() (int64, error) {
	cutoff := time.Now().AddDate(0, -6, 0)
	result := utils.DB.Where("created_at < ?", cutoff).Unscoped().Delete(&models.LoginLog{})
	return result.RowsAffected, result.Error
}
