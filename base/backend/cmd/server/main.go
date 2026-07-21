package main

import (
	"base/config"
	"base/internal/adapter"
	"base/internal/routes"
	"base/internal/seed"
	"base/pkg/db"
	"base/pkg/redis"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		logrus.WithError(err).Fatal("加载配置失败")
	}

	if err := db.Init(&cfg.MySQL); err != nil {
		logrus.WithError(err).Fatal("初始化数据库失败")
	}

	if err := redis.Init(&cfg.Redis); err != nil {
		logrus.WithError(err).Fatal("初始化Redis失败")
	}

	if err := adapter.DefaultRegistry.Reload(); err != nil {
		logrus.WithError(err).Warn("加载应用注册表失败")
	}

	if err := seed.Run(); err != nil {
		logrus.WithError(err).Fatal("初始化默认数据失败")
	}

	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()
	routes.Register(r)

	logrus.Infof("Base server listening on :%s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		logrus.WithError(err).Fatal("启动服务失败")
	}
}
