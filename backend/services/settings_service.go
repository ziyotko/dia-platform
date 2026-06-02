package services

import (
	"errors"

	"server/models"
	"server/utils"

	"gorm.io/gorm"
)

type SettingsService struct{}

func (s *SettingsService) GetSettings() (*models.Settings, error) {
	var settings models.Settings
	result := utils.DB.First(&settings)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			settings = models.Settings{
				SiteName:          "门户网站管理后台",
				Icp:               "京ICP备12345678号",
				Copyright:         "门户网站管理系统 版权所有",
				CaptchaEnabled:    true,
				LockEnabled:       true,
				MaxFailCount:      5,
				LockDuration:      30,
				MinPasswordLength: 8,
				TokenExpire:       24,
				SmtpHost:          "smtp.example.com",
				SmtpPort:          "587",
				FromEmail:         "noreply@example.com",
				FromName:          "系统通知",
				Ssl:               true,
				ThemeColor:        "#409eff",
				SidebarStyle:      "light",
				TagsView:          true,
				Breadcrumb:        true,
			}
			if err := utils.DB.Create(&settings).Error; err != nil {
				return nil, err
			}
			return &settings, nil
		}
		return nil, result.Error
	}
	return &settings, nil
}

func (s *SettingsService) UpdateSettings(settings *models.Settings) error {
	var existing models.Settings
	result := utils.DB.First(&existing)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return utils.DB.Create(settings).Error
		}
		return result.Error
	}
	settings.ID = existing.ID
	return utils.DB.Model(&existing).Updates(settings).Error
}
