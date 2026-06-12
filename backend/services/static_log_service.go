package services

import (
	"server/models"
	"server/utils"
)

type StaticLogService struct{}

type StaticLogListResult struct {
	Total int64             `json:"total"`
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

func (s *StaticLogService) Create(log *models.StaticLog) error {
	return utils.DB.Create(log).Error
}
