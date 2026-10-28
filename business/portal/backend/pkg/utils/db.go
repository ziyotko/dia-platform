package utils

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"portal/config"
)

type gormLogWriter struct{}

func (w *gormLogWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))
	if msg != "" {
		Logger.Info(msg)
	}
	return len(p), nil
}

// gormLogLevel 让 GORM 的 SQL 日志跟随 log.level：
// debug 时输出全部 SQL；其它级别只输出慢查询与错误。
// 否则启动/迁移阶段的数百条 SQL 都要格式化并写盘（lumberjack 还会压缩归档），白白拖慢启动。
func gormLogLevel(level string) logger.LogLevel {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug", "trace":
		return logger.Info
	case "warn", "warning":
		return logger.Warn
	case "error", "fatal", "panic":
		return logger.Error
	}
	return logger.Warn
}

// dbPingTimeout 启动时探测数据库的等待上限：连不上应尽快失败，而不是卡在 DSN 的超时上。
const dbPingTimeout = 10 * time.Second

var DB *gorm.DB

func InitDB() {
	conf := config.AppConfig.MySQL

	parseTime := "false"
	if conf.ParseTime {
		parseTime = "true"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%s&loc=%s&timeout=%s&readTimeout=%s&writeTimeout=%s",
		conf.User,
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
			SlowThreshold: time.Second,
			LogLevel:      gormLogLevel(config.AppConfig.Log.Level),
			// 参数化输出 SQL（占位符 + 参数分列）：默认模式会把参数插值进 SQL 文本，
			// 导致 SMTP 授权码、密码哈希、邮箱/手机号被明文写入 logs/*.log（保留 180 天）。
			ParameterizedQueries:      true,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
		// 单条写操作不再包一层事务（播种/清理等单语句场景无需事务开销）
		SkipDefaultTransaction: true,
		// 下方显式 PingContext，避免 gorm.Open 再建连握手一次
		DisableAutomaticPing: true,
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

	// 连接池参数在 Ping 之前设置：Ping 建立的连接直接进入正确的池配置
	if conf.MaxIdle > 0 {
		sqlDB.SetMaxIdleConns(conf.MaxIdle)
	} else {
		sqlDB.SetMaxIdleConns(10)
	}

	if conf.MaxOpen > 0 {
		sqlDB.SetMaxOpenConns(conf.MaxOpen)
	} else {
		sqlDB.SetMaxOpenConns(100)
	}

	if conf.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetime) * time.Second)
	}

	if conf.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(conf.ConnMaxIdleTime) * time.Second)
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbPingTimeout)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %s", err)
	}
}
