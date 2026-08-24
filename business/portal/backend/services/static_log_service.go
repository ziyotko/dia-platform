package services

import (
	"server/models"
	"server/utils"
)

type StaticLogService struct{}

type StaticLogListResult struct {
	Total int64              `json:"total"`
	List  []models.StaticLog `json:"list"`
}

func (s *StaticLogService) GetList(page, pageSize int) (*StaticLogListResult, error) {
	var logs []models.StaticLog
	var total int64

	query := utils.DB.Model(&models.StaticLog{})

	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	err = query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	if err != nil {
		return nil, err
	}

	return &StaticLogListResult{
		Total: total,
		List:  logs,
	}, nil
}

func (s *StaticLogService) Clear() error {
	return utils.DB.Where("1 = 1").Unscoped().Delete(&models.StaticLog{}).Error
}

// CreateIfNotExists 按 任务ID + 状态 去重创建静态化日志（同一任务同一状态只记录一次）
func (s *StaticLogService) CreateIfNotExists(log *models.StaticLog) error {
	if log.JobID != "" {
		var count int64
		if err := utils.DB.Model(&models.StaticLog{}).
			Where("job_id = ? AND status = ?", log.JobID, log.Status).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
	}
	return utils.DB.Create(log).Error
}
