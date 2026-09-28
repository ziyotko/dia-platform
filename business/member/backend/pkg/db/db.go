package db

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"member/config"
	"member/pkg/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// gormLogWriter 把 GORM 的 SQL 日志接到应用日志（与 portal 一致）：
// 否则 SQL 只写 stdout，不会进 logs/<name>.log，排障时拿不到历史。
type gormLogWriter struct{}

func (w *gormLogWriter) Write(p []byte) (int, error) {
	msg := strings.TrimSpace(string(p))
	if msg != "" {
		utils.LogInfo("%s", msg)
	}
	return len(p), nil
}

func Init(cfg *config.MySQLConfig) {
	// 时区：优先用配置的 loc（推荐显式 Asia/Shanghai，不依赖部署机时区），留空回退 Local。
	loc := cfg.Loc
	if loc == "" {
		loc = "Local"
	}
	params := "charset=" + cfg.Charset + "&parseTime=True&loc=" + url.QueryEscape(loc)
	if cfg.Timeout != "" {
		params += "&timeout=" + cfg.Timeout
	}
	if cfg.ReadTimeout != "" {
		params += "&readTimeout=" + cfg.ReadTimeout
	}
	if cfg.WriteTimeout != "" {
		params += "&writeTimeout=" + cfg.WriteTimeout
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, params)

	// 参数化输出 SQL（占位符 + 参数分列）：默认模式会把参数插值进 SQL 文本，
	// 使得密码哈希 / 手机号 / 证件号等明文落进日志文件（与 portal 一致）。
	newLogger := logger.New(
		log.New(&gormLogWriter{}, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			ParameterizedQueries:      true,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:                                   newLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connection established")
}
