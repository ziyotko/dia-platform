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
	TrustedProxies   []string `mapstructure:"trusted_proxies"`
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

	if Cfg.MySQL.Password == "" || Cfg.MySQL.Password == "APPLICATION_DB_PASSWORD" {
		log.Printf("[WARN] 未设置数据库密码！必须设置环境变量 APPLICATION_DB_PASSWORD。")
		log.Fatal("程序退出")
	}

	if Cfg.JWT.Secret == "" || Cfg.JWT.Secret == "APPLICATION_JWT_SECRET" {
		log.Printf("[WARN] 未设置 JWT 密钥！必须设置环境变量 APPLICATION_JWT_SECRET。")
		log.Fatal("程序退出")
	}
}
