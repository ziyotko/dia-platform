package utils

import (
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"

	"portal/config"
)

var Logger *logrus.Logger

func InitLogger() {
	Logger = logrus.New()

	logLevel, err := logrus.ParseLevel(config.AppConfig.Log.Level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	Logger.SetLevel(logLevel)

	logPath := config.AppConfig.Log.Path
	logFileName := filepath.Join(logPath, time.Now().Format("2006-01-02")+".log")

	// 轮转参数统一走配置（与 member / application 同名字段），未配置时回退历史默认值
	maxSize := config.AppConfig.Log.MaxSize
	if maxSize <= 0 {
		maxSize = 100
	}
	maxBackups := config.AppConfig.Log.MaxBackups
	if maxBackups <= 0 {
		maxBackups = 1000
	}
	maxAge := config.AppConfig.Log.MaxAge
	if maxAge <= 0 {
		maxAge = 180
	}

	Logger.SetOutput(&lumberjack.Logger{
		Filename:   logFileName,
		MaxSize:    maxSize,    // 单个文件最大大小（MB）
		MaxBackups: maxBackups, // 保留的旧日志文件最大数量
		MaxAge:     maxAge,     // 保留日志文件的最大天数
		Compress:   true,       // 是否压缩旧日志文件
	})

	Logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
}
