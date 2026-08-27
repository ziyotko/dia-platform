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
	Port             string   `mapstructure:"port"`
	Host             string   `mapstructure:"host"`
	Mode             string   `mapstructure:"mode"`
	ApiPrefix        string   `mapstructure:"api_prefix"`
	UploadDirPrefix  string   `mapstructure:"upload_dir_prefix"`
	MaxConcurrentIPs int      `mapstructure:"max_concurrent_ips"`
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
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
