package utils

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"server/config"
)

type gormLogWriter struct{}

func (w *gormLogWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))
	if msg != "" {
		Logger.Info(msg)
	}
	return len(p), nil
}

var DB *gorm.DB

func InitDB() {
	conf := config.AppConfig.Database

	parseTime := "false"
	if conf.ParseTime {
		parseTime = "true"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%s&loc=%s&timeout=%s&readTimeout=%s&writeTimeout=%s",
		conf.Username,
		conf.Password,
		conf.Host,
		conf.Port,
		conf.DBName,
		conf.Charset,
		parseTime,
		url.QueryEscape(conf.Loc),
		conf.Timeout,
		conf.ReadTimeout,
		conf.WriteTimeout,
	)
	var err error
	newLogger := logger.New(
		log.New(&gormLogWriter{}, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		log.Fatalf("Failed to connect database: %s", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to get DB: %s", err)
	}

	err = sqlDB.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %s", err)
	}

	if conf.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(conf.MaxIdleConns)
	} else {
		sqlDB.SetMaxIdleConns(10)
	}

	if conf.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(conf.MaxOpenConns)
	} else {
		sqlDB.SetMaxOpenConns(100)
	}

	if conf.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetime) * time.Second)
	}

	if conf.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(conf.ConnMaxIdleTime) * time.Second)
	}
}
