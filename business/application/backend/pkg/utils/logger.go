package utils

import (
	"io"
	"os"

	"application/config"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *logrus.Logger

func InitLogger() {
	cfg := config.Cfg.Log
	Logger = logrus.New()

	Logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	lumberjackLogger := &lumberjack.Logger{
		Filename:   cfg.Path,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   true,
	}

	multiWriter := io.MultiWriter(os.Stdout, lumberjackLogger)
	Logger.SetOutput(multiWriter)
	// 日志级别走配置（与 portal / member 一致），非法值或未配置回退 info
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	Logger.SetLevel(level)
}
