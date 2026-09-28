package utils

import (
	"fmt"
	"io"
	"member/config"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *logrus.Logger

// LogInfo 安全写信息日志：Logger 未初始化（测试/脚本环境）时退化为标准输出，避免 nil panic。
// 供 DB 层（GORM SQL 日志）等不直接持有 Logger 的包使用。
func LogInfo(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Infof(format, args...)
		return
	}
	fmt.Printf("[INFO] "+format+"\n", args...)
}

// LogWarn 安全写告警日志：Logger 未初始化（测试/脚本环境）时退化为标准输出，避免 nil panic。
func LogWarn(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Warnf(format, args...)
		return
	}
	fmt.Printf("[WARN] "+format+"\n", args...)
}

func InitLogger() {
	cfg := config.Cfg.Log
	Logger = logrus.New()

	if cfg.Path != "" {
		ljWriter := &lumberjack.Logger{
			Filename:   cfg.Path,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   true,
		}
		multiWriter := io.MultiWriter(os.Stdout, ljWriter)
		Logger.SetOutput(multiWriter)
	} else {
		Logger.SetOutput(os.Stdout)
	}

	Logger.SetFormatter(&logrus.TextFormatter{
		// 与 portal / application 统一时间格式
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
	// 日志级别走配置（与 portal / application 一致），非法值或未配置回退 info
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	Logger.SetLevel(level)
}
