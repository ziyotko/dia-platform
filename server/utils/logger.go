package utils

import (
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"

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
	if err := os.MkdirAll(logPath, 0755); err != nil {
		panic(err)
	}

	fileName := filepath.Join(logPath, time.Now().Format("2006-01-02")+".log")
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

	Logger.SetOutput(file)

	Logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
}
