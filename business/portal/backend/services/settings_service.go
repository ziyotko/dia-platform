package services

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strings"
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

				Copyright:              "门户网站管理系统 版权所有",
				CaptchaEnabled:         true,
				LockEnabled:            true,
				MaxFailCount:           5,
				LockDuration:           30,
				MinPasswordLength:      8,
				TokenExpire:            24,
				SmtpHost:               "smtp.example.com",
				SmtpPort:               "587",
				FromEmail:              "noreply@example.com",
				FromName:               "系统通知",
				Ssl:                    true,
				StaticPath:             "",
				HomeGray:               false,
				StaticProgramAddr:      "",
				StaticProgramTokenName: "",
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

// TestEmailConnection 测试邮件（SMTP）连接：按当前设置建立连接、按需 TLS/STARTTLS 并尝试认证，
// 不发送任何邮件。返回的 error 包含具体失败原因，便于管理员定位配置问题。
func (s *SettingsService) TestEmailConnection() error {
	settings, err := s.GetSettings()
	if err != nil {
		return errors.New("获取系统设置失败")
	}

	host := strings.TrimSpace(settings.SmtpHost)
	port := strings.TrimSpace(settings.SmtpPort)
	if host == "" || port == "" {
		return errors.New("SMTP 服务器地址或端口未配置")
	}

	addr := net.JoinHostPort(host, port)
	const timeout = 10 * time.Second

	var conn net.Conn
	// 465 端口为隐式 TLS（SSL），其余端口在启用 SSL 时使用 STARTTLS
	implicitTLS := settings.Ssl && port == "465"
	if implicitTLS {
		conn, err = tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", addr, &tls.Config{ServerName: host})
	} else {
		conn, err = net.DialTimeout("tcp", addr, timeout)
	}
	if err != nil {
		return fmt.Errorf("无法连接 SMTP 服务器 %s: %w", addr, err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("初始化 SMTP 客户端失败: %w", err)
	}
	defer client.Close()

	if settings.Ssl && !implicitTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: host}); err != nil {
				return fmt.Errorf("STARTTLS 握手失败: %w", err)
			}
		}
	}

	if settings.EmailPassword != "" {
		auth := smtp.PlainAuth("", strings.TrimSpace(settings.FromEmail), settings.EmailPassword, host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP 认证失败: %w", err)
		}
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("断开 SMTP 连接失败: %w", err)
	}
	return nil
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
