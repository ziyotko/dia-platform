package services

import (
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

func (s *LogService) ClearLogs() error {
	return utils.DB.Where("1 = 1").Delete(&models.OperationLog{}).Error
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
	Total int64               `json:"total"`
	List  []models.LoginLog   `json:"list"`
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

func (s *LogService) GetRecentLoginLogs(limit int) ([]models.LoginLog, error) {
	var logs []models.LoginLog
	err := utils.DB.Model(&models.LoginLog{}).Order("id DESC").Limit(limit).Find(&logs).Error
	return logs, err
}
