package config

import (
	"log"

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
	ApiPrefix        string `mapstructure:"api_prefix"`
	UploadDirPrefix  string `mapstructure:"upload_dir_prefix"`
	MaxConcurrentIPs int    `mapstructure:"max_concurrent_ips"`
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
}
