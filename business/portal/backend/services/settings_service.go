package services

import (
	"errors"

	"server/models"
	"server/utils"

	"gorm.io/gorm"
)

type SettingsService struct{}

func (s *SettingsService) GetSettings() (*models.Setting, error) {
	var settings models.Setting
	result := utils.DB.First(&settings)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			settings = models.Setting{
				SiteName:                 "门户网站管理后台",
				Icp:                      "京ICP备12345678号",
				Copyright:                "门户网站管理系统 版权所有",
				CaptchaEnabled:           true,
				LockEnabled:              true,
				MaxFailCount:             5,
				LockDuration:             30,
				MinPasswordLength:        8,
				TokenExpire:              24,
				SmtpHost:                 "smtp.example.com",
				SmtpPort:                 "587",
				FromEmail:                "noreply@example.com",
				FromName:                 "系统通知",
				Ssl:                      true,
				ThemeColor:               "#409eff",
				SidebarStyle:             "light",
				TagsView:                 true,
				Breadcrumb:               true,
				HomeGray:                 false,
				HomeStaticTimeEnabled:    false,
				HomeStaticTime:           "",
				ColumnStaticTimeEnabled:  false,
				ColumnStaticTime:         "",
				SpecialStaticTimeEnabled: false,
				SpecialStaticTime:        "",
				DetailStaticTimeEnabled:  false,
				DetailStaticTime:         "",
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

// UpdateSettings 只更新 fields 中指定的字段，避免请求中未携带的字段（零值）覆盖表中其他设置项
func (s *SettingsService) UpdateSettings(settings *models.Setting, fields ...string) error {
	if len(fields) == 0 {
		return nil
	}
	var existing models.Setting
	result := utils.DB.First(&existing)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return utils.DB.Create(settings).Error
		}
		return result.Error
	}
	settings.ID = existing.ID
	return utils.DB.Model(&existing).Select(fields).Updates(settings).Error
}
