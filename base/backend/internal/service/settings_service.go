package service

import (
	"strconv"

	"base/internal/models"
	"base/pkg/db"

	"gorm.io/gorm"
)

type SettingsService struct{}

// SecuritySettings 安全策略（系统设置 → 安全策略），由登录与密码相关流程读取。
type SecuritySettings struct {
	CaptchaEnabled bool // 是否启用登录验证码
	LockEnabled    bool // 是否启用登录失败锁定
	MaxFailCount   int  // 最大失败次数（超过则锁定）
	LockDuration   int  // 锁定时长（分钟）
	PwdMinLength   int  // 密码最小长度
}

// GetSecuritySettings 读取安全策略，未配置项使用默认值（验证码开启、5 次锁 30 分钟、密码至少 8 位）。
func (s SettingsService) GetSecuritySettings() SecuritySettings {
	out := SecuritySettings{
		CaptchaEnabled: true,
		LockEnabled:    true,
		MaxFailCount:   5,
		LockDuration:   30,
		PwdMinLength:   8,
	}
	raw, err := s.GetByCategory("security")
	if err != nil {
		return out
	}
	out.CaptchaEnabled = boolValue(raw["captchaEnabled"], out.CaptchaEnabled)
	out.LockEnabled = boolValue(raw["loginLock"], out.LockEnabled)
	out.MaxFailCount = intValue(raw["maxFailCount"], out.MaxFailCount, 3, 20)
	out.LockDuration = intValue(raw["lockDuration"], out.LockDuration, 1, 1440)
	out.PwdMinLength = intValue(raw["pwdMinLength"], out.PwdMinLength, 6, 32)
	return out
}

func boolValue(raw string, fallback bool) bool {
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return v
}

func intValue(raw string, fallback, min, max int) int {
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < min || v > max {
		return fallback
	}
	return v
}

func (s SettingsService) GetAll() (map[string]map[string]string, error) {
	var list []models.Setting
	if err := db.DB.Find(&list).Error; err != nil {
		return nil, err
	}
	result := make(map[string]map[string]string)
	for _, item := range list {
		if result[item.Category] == nil {
			result[item.Category] = make(map[string]string)
		}
		result[item.Category][item.Key] = item.Value
	}
	return result, nil
}

func (s SettingsService) GetByCategory(category string) (map[string]string, error) {
	var list []models.Setting
	if err := db.DB.Where("category = ?", category).Find(&list).Error; err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, item := range list {
		result[item.Key] = item.Value
	}
	return result, nil
}

func (s SettingsService) BatchSave(items []models.Setting) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if item.Category == "" || item.Key == "" {
				continue
			}
			var existing models.Setting
			err := tx.Where("category = ? AND key = ?", item.Category, item.Key).First(&existing).Error
			if err == nil {
				existing.Value = item.Value
				existing.Type = item.Type
				existing.Remark = item.Remark
				if err := tx.Save(&existing).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Create(&item).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
