package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server Server `mapstructure:"server"`
	MySQL  MySQL  `mapstructure:"mysql"`
	Redis  Redis  `mapstructure:"redis"`
	JWT    JWT    `mapstructure:"jwt"`
}

type Server struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`

	// 上传目录与单文件大小上限（MB）
	UploadDir   string `mapstructure:"upload_dir"`
	MaxUploadMB int    `mapstructure:"max_upload_mb"`

	// 可信反向代理地址，决定 gin 是否信任 X-Forwarded-For / X-Real-IP（否则一律使用 RemoteAddr，防伪造头绕过限流）
	TrustedProxies []string `mapstructure:"trusted_proxies"`

	// 登录接口限流（防暴力破解，按真实客户端 IP）
	LoginRateLimit      int `mapstructure:"login_rate_limit"`          // 每窗口允许的最大请求数
	LoginRateWindowSecs int `mapstructure:"login_rate_window_seconds"` // 限流窗口（秒）

	// 验证码接口限流（防刷验证码，按真实客户端 IP）
	CaptchaRateLimit      int `mapstructure:"captcha_rate_limit"`
	CaptchaRateWindowSecs int `mapstructure:"captcha_rate_window_seconds"`

	// 初始化超管接口限流（公开写接口，防被反复调用）
	InitRateLimit      int `mapstructure:"init_rate_limit"`
	InitRateWindowSecs int `mapstructure:"init_rate_window_seconds"`

	// 工作流超时提醒的后台扫描间隔（秒），<= 0 表示关闭
	WorkflowRemindInterval int `mapstructure:"workflow_remind_interval_seconds"`
}

type MySQL struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	Charset  string `mapstructure:"charset"`
	MaxOpen  int    `mapstructure:"max_open"`
	MaxIdle  int    `mapstructure:"max_idle"`
}

type Redis struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWT struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
	Issuer      string `mapstructure:"issuer"`
}

var Cfg *Config

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	cfg.applyDefaults()
	Cfg = &cfg
	return &cfg, nil
}

// applyDefaults 为未配置的项补充默认值（兼容旧版 config.yaml，未新增限流配置时也能正常工作）。
func (c *Config) applyDefaults() {
	if c.Server.Mode == "" {
		c.Server.Mode = "release"
	}
	if len(c.Server.TrustedProxies) == 0 {
		c.Server.TrustedProxies = []string{"127.0.0.1"}
	}

	setRateLimit(&c.Server.LoginRateLimit, &c.Server.LoginRateWindowSecs, 10, 60)
	setRateLimit(&c.Server.CaptchaRateLimit, &c.Server.CaptchaRateWindowSecs, 30, 60)
	setRateLimit(&c.Server.InitRateLimit, &c.Server.InitRateWindowSecs, 5, 60)

	if c.Server.UploadDir == "" {
		c.Server.UploadDir = "./uploads"
	}
	if c.Server.MaxUploadMB <= 0 {
		c.Server.MaxUploadMB = 50
	}
	// 工作流超时提醒：默认 10 分钟扫一次；显式配置 0 或负数表示关闭
	if c.Server.WorkflowRemindInterval == 0 {
		c.Server.WorkflowRemindInterval = 600
	}
}

func setRateLimit(limit, windowSecs *int, defaultLimit, defaultWindowSecs int) {
	if *limit <= 0 {
		*limit = defaultLimit
	}
	if *windowSecs <= 0 {
		*windowSecs = defaultWindowSecs
	}
}
