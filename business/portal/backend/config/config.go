package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Log      LogConfig
}

type ServerConfig struct {
	Port             string `mapstructure:"port"`
	Host             string `mapstructure:"host"`
	Mode             string `mapstructure:"mode"`
	ApiPrefix        string `mapstructure:"api_prefix"`
	UploadDirPrefix  string `mapstructure:"upload_dir_prefix"`
	MaxConcurrentIPs int    `mapstructure:"max_concurrent_ips"`
	// 非 multipart（JSON）请求体上限（MB）：富文本正文会内联 base64 图片，需留足余量；
	// <=0 时回退 64MB。文件上传（multipart）不受此限制。
	MaxJSONBodyMB  int      `mapstructure:"max_json_body_mb"`
	AllowedOrigins []string `mapstructure:"allowed_origins"`

	// 可信反向代理地址，决定 gin 是否信任 X-Forwarded-For / X-Real-IP
	TrustedProxies []string `mapstructure:"trusted_proxies"`

	// 站点分析等公开写接口的限流配置（按真实客户端 IP）
	AnalyticsRateLimit      int `mapstructure:"analytics_rate_limit"`          // 每窗口允许的最大请求数
	AnalyticsRateWindowSecs int `mapstructure:"analytics_rate_window_seconds"` // 限流窗口（秒）

	// 登录接口限流（防暴力破解，按真实客户端 IP）
	LoginRateLimit      int `mapstructure:"login_rate_limit"`          // 每窗口允许的最大请求数
	LoginRateWindowSecs int `mapstructure:"login_rate_window_seconds"` // 限流窗口（秒）

	// 验证码接口限流（防刷验证码，按真实客户端 IP）
	CaptchaRateLimit      int `mapstructure:"captcha_rate_limit"`          // 每窗口允许的最大请求数
	CaptchaRateWindowSecs int `mapstructure:"captcha_rate_window_seconds"` // 限流窗口（秒）

	// 公开只读接口限流（/site-info、/search/articles 等无需认证的查询接口，按真实客户端 IP）
	PublicRateLimit      int `mapstructure:"public_rate_limit"`          // 每窗口允许的最大请求数（<=0 时回退 60）
	PublicRateWindowSecs int `mapstructure:"public_rate_window_seconds"` // 限流窗口（秒，<=0 时回退 60）

	// 防重放/请求签名校验（主动检测与临时封禁）
	ReplayWindowSecs int `mapstructure:"replay_window_seconds"` // 时间戳新鲜度窗口（秒），默认 120
	ReplayMaxFail    int `mapstructure:"replay_max_fail"`       // 触发临时封禁的失败次数
	ReplayBanMinutes int `mapstructure:"replay_ban_minutes"`    // 临时封禁时长（分钟）
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            string `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	Charset         string `mapstructure:"charset"`
	ParseTime       bool   `mapstructure:"parse_time"`
	Loc             string `mapstructure:"loc"`
	Timeout         string `mapstructure:"timeout"`
	ReadTimeout     string `mapstructure:"read_timeout"`
	WriteTimeout    string `mapstructure:"write_timeout"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime int    `mapstructure:"conn_max_idle_time"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	DB1      int    `mapstructure:"db1"`
	DB2      int    `mapstructure:"db2"`
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpiresHour int    `mapstructure:"expires_hour"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
	Path  string `mapstructure:"path"`
}

var AppConfig Config

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Unable to decode config: %s", err)
	}

	applyEnvOverrides()
}

// applyEnvOverrides 允许通过环境变量覆盖敏感配置，避免把密钥/密码提交到仓库。
// 生产环境务必注入：PORTAL_JWT_SECRET、PORTAL_DB_PASSWORD。
func applyEnvOverrides() {

	if v := os.Getenv("PORTAL_JWT_SECRET"); v != "" {
		AppConfig.JWT.Secret = v
	}

	if v := os.Getenv("PORTAL_DB_PASSWORD"); v != "" {
		AppConfig.Database.Password = v
	}

	if AppConfig.JWT.Secret == "" || AppConfig.JWT.Secret == "PORTAL_JWT_SECRET" {
		log.Printf("[WARN] 未设置 JWT 密钥！必须设置环境变量 PORTAL_JWT_SECRET。")
		log.Fatal("程序退出")
	}

	if AppConfig.Database.Password == "" || AppConfig.Database.Password == "PORTAL_DB_PASSWORD" {
		log.Printf("[WARN] 未设置数据库密码！必须设置环境变量 PORTAL_DB_PASSWORD。")
		log.Fatal("程序退出")
	}
}
