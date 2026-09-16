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
		FullTimestamp: true,
	})
	Logger.SetLevel(logrus.InfoLevel)
}
