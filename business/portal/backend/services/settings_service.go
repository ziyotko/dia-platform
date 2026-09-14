package services

import (
	"encoding/json"
	"errors"
	"time"

	"server/models"
	"server/utils"

	"gorm.io/gorm"
)

type SettingsService struct{}

func (s *SettingsService) GetMinPasswordLengthSettings() (*models.Setting, error) {
	var settings models.Setting
	result := utils.DB.Select("min_password_length").First(&settings)
	if result.Error != nil {
		return nil, result.Error
	}
	return &settings, nil
}

func (s *SettingsService) GetSettings() (*models.Setting, error) {
	var settings models.Setting
	result := utils.DB.First(&settings)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			settings = models.Setting{
				SiteName: "门户网站管理后台",
				SiteUrl:  "",
				Icp:      "京ICP备12345678号",

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
				StaticPath:               "",
				HomeGray:                 false,
				HomeStaticTimeEnabled:    false,
				HomeStaticTime:           "",
				ColumnStaticTimeEnabled:  false,
				ColumnStaticTime:         "",
				SpecialStaticTimeEnabled: false,
				SpecialStaticTime:        "",
				DetailStaticTimeEnabled:  false,
				DetailStaticTime:         "",
				StaticProgramAddr:        "",
				StaticProgramTokenName:   "",
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

// StaticParams 静态化相关参数
type StaticParams struct {
	StaticPath             string `json:"staticPath"`             // 静态化输出路径
	StaticProgramAddr      string `json:"staticProgramAddr"`      // 静态化程序访问地址
	StaticProgramTokenName string `json:"staticProgramTokenName"` // 静态化程序访问令牌名
	HomeGray               bool   `json:"homeGray"`               // 首页整体变灰
}

// StaticParamsCacheKey 静态化参数在缓存（Redis2）中的 Key
const StaticParamsCacheKey = "db:static:params"

// staticParamsCacheTTL 静态化参数缓存有效期。
// 设置 TTL 可保证即使某次刷新失败，缓存也会过期回源，避免 DB 与缓存长期不一致。
const staticParamsCacheTTL = 24 * time.Hour

// GetStaticParams 从数据库读取静态化相关参数
func (s *SettingsService) GetStaticParams() (*StaticParams, error) {
	settings, err := s.GetSettings()
	if err != nil {
		return nil, err
	}
	return &StaticParams{
		StaticPath:             settings.StaticPath,
		StaticProgramAddr:      settings.StaticProgramAddr,
		StaticProgramTokenName: settings.StaticProgramTokenName,
		HomeGray:               settings.HomeGray,
	}, nil
}

// LoadStaticParamsToCache 从数据库读取静态化参数并写入缓存（后端启动时调用）
func (s *SettingsService) LoadStaticParamsToCache() error {
	params, err := s.GetStaticParams()
	if err != nil {
		return err
	}
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	return utils.Redis2.Set(utils.Ctx, StaticParamsCacheKey, data, staticParamsCacheTTL).Err()
}

// InvalidateStaticParamsCache 删除静态化参数缓存，下次读取时将回源数据库。
func (s *SettingsService) InvalidateStaticParamsCache() error {
	return utils.Redis2.Del(utils.Ctx, StaticParamsCacheKey).Err()
}

// GetStaticParamsFromCache 从缓存读取静态化参数；缓存未命中或解析失败时回源数据库并刷新缓存
func (s *SettingsService) GetStaticParamsFromCache() (*StaticParams, error) {
	if data, err := utils.Redis2.Get(utils.Ctx, StaticParamsCacheKey).Bytes(); err == nil {
		var params StaticParams
		if json.Unmarshal(data, &params) == nil {
			return &params, nil
		}
	}
	// 缓存未命中或解析失败：回源数据库并刷新缓存
	params, err := s.GetStaticParams()
	if err != nil {
		return nil, err
	}
	_ = s.LoadStaticParamsToCache()
	return params, nil
}
