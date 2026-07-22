package utils

import (
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"

	"server/config"
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

	Logger.SetOutput(&lumberjack.Logger{
		Filename:   logFileName,
		MaxSize:    100,  // 单个文件最大大小（MB）
		MaxBackups: 1000, // 保留的旧日志文件最大数量
		MaxAge:     180,  // 保留日志文件的最大天数
		Compress:   true, // 是否压缩旧日志文件
	})

	Logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
}
