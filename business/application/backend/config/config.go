package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

var Cfg *Config

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	Redis  RedisConfig  `mapstructure:"redis"`
	JWT    JWTConfig    `mapstructure:"jwt"`
	Log    LogConfig    `mapstructure:"log"`
}

type ServerConfig struct {
	Port             int      `mapstructure:"port"`
	Mode             string   `mapstructure:"mode"`
	APIPrefix        string   `mapstructure:"api_prefix"`
	MaxConcurrentIPs int      `mapstructure:"max_concurrent_ips"`
	UploadDirPrefix  string   `mapstructure:"upload_dir_prefix"`
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	TrustedProxies   []string `mapstructure:"trusted_proxies"`

	// 非 multipart（JSON）请求体上限（MB）：<=0 回退 64。文件上传（multipart）不受此限制。
	MaxJSONBodyMB int `mapstructure:"max_json_body_mb"`

	// 固定窗口限流（次数/分钟，按真实客户端 IP；<=0 回退代码默认值）。
	// 与 portal / member 同名同义：登录、验证码、上传。
	LoginRateLimit   int `mapstructure:"login_rate_limit"`
	CaptchaRateLimit int `mapstructure:"captcha_rate_limit"`
	UploadRateLimit  int `mapstructure:"upload_rate_limit"`
}

type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
	Charset  string `mapstructure:"charset"`
	MaxOpen  int    `mapstructure:"max_open"`
	MaxIdle  int    `mapstructure:"max_idle"`

	// 时区与超时（与 portal / member 的 mysql 段同名同义）：
	// loc 留空用 Local（部署机时区必须是 Asia/Shanghai）；timeout/read_timeout/write_timeout 留空则不附加。
	Loc          string `mapstructure:"loc"`
	Timeout      string `mapstructure:"timeout"`
	ReadTimeout  string `mapstructure:"read_timeout"`
	WriteTimeout string `mapstructure:"write_timeout"`
}

type RedisConfig struct {
	Addr         string `mapstructure:"addr"`
	Password     string `mapstructure:"password"`
	CaptchaDB    int    `mapstructure:"captcha_db"`
	AntiReplayDB int    `mapstructure:"anti_replay_db"`
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
	Issuer      string `mapstructure:"issuer"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Path       string `mapstructure:"path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

func Load(path string) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic("Failed to read config: " + err.Error())
	}
	Cfg = &Config{}
	if err := viper.Unmarshal(Cfg); err != nil {
		panic("Failed to unmarshal config: " + err.Error())
	}
	applyEnvOverrides()
}

// applyEnvOverrides 用环境变量覆盖敏感配置，避免把数据库密码 / JWT 密钥提交到仓库。
// 生产环境务必注入：APPLICATION_DB_PASSWORD、APPLICATION_JWT_SECRET。
func applyEnvOverrides() {
	if v := os.Getenv("APPLICATION_DB_PASSWORD"); v != "" {
		Cfg.MySQL.Password = v
	}
	if v := os.Getenv("APPLICATION_JWT_SECRET"); v != "" {
		Cfg.JWT.Secret = v
	}
	if v := os.Getenv("APPLICATION_MODE"); v != "" {
		Cfg.Server.Mode = v
	}

	// 与 portal / member 一致：mode 留空时回退 release，避免默认输出调试信息
	if Cfg.Server.Mode == "" {
		Cfg.Server.Mode = "release"
	}

	if Cfg.MySQL.Password == "" || Cfg.MySQL.Password == "APPLICATION_DB_PASSWORD" {
		log.Printf("[WARN] 未设置数据库密码！必须设置环境变量 APPLICATION_DB_PASSWORD。")
		log.Fatal("程序退出")
	}

	if Cfg.JWT.Secret == "" || Cfg.JWT.Secret == "APPLICATION_JWT_SECRET" {
		log.Printf("[WARN] 未设置 JWT 密钥！必须设置环境变量 APPLICATION_JWT_SECRET。")
		log.Fatal("程序退出")
	}
}
