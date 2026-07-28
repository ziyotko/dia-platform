package utils

import (
	"io"
	"member/config"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *logrus.Logger

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
		FullTimestamp: true,
	})
	Logger.SetLevel(logrus.InfoLevel)
}
