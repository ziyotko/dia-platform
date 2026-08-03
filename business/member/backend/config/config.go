package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

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
}

// 已知的弱默认密钥，用于启动时告警
const defaultJWTSecret = "member-jwt-secret-key-2024"

type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
	Charset  string `mapstructure:"charset"`
	MaxOpen  int    `mapstructure:"max_open"`
	MaxIdle  int    `mapstructure:"max_idle"`
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
	Path       string `mapstructure:"path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

var Cfg *Config

func Load(path string) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}
	Cfg = &Config{}
	if err := viper.Unmarshal(Cfg); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}
	applyEnvOverrides()
}

// applyEnvOverrides 允许通过环境变量覆盖敏感配置，避免把密钥/密码提交到仓库。
// 生产环境务必通过环境变量注入：MEMBER_JWT_SECRET、MEMBER_DB_PASSWORD、MEMBER_MODE。
func applyEnvOverrides() {
	if v := os.Getenv("MEMBER_JWT_SECRET"); v != "" {
		Cfg.JWT.Secret = v
	}
	if v := os.Getenv("MEMBER_DB_PASSWORD"); v != "" {
		Cfg.MySQL.Password = v
	}
	if v := os.Getenv("MEMBER_MODE"); v != "" {
		Cfg.Server.Mode = v
	}

	if Cfg.Server.Mode == "" {
		Cfg.Server.Mode = "release"
	}

	if Cfg.JWT.Secret == "" || Cfg.JWT.Secret == defaultJWTSecret {
		log.Printf("[WARN] 正在使用弱/默认 JWT 密钥！生产环境必须设置环境变量 MEMBER_JWT_SECRET。")
	}
}
