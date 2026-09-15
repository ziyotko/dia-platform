package main

import (
	"flag"

	"base/config"
	"base/internal/adapter"
	"base/internal/routes"
	"base/internal/seed"
	"base/internal/service"
	"base/pkg/db"
	"base/pkg/notifier"
	"base/pkg/redis"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	mockData := flag.Bool("mock-data", false, "插入模拟数据（仅本地测试）")
	flag.Parse()

	cfg, err := config.Load("config.yaml")
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

	// 平台超级管理员（唯一实现）：幂等创建，必须在 seed.Run 之前——
	// seed 会把菜单授予该 admin 账号（seedSuperAdminRole 依赖它已存在）
	if created, err := (service.AuthService{}).EnsureSuperAdmin("admin123"); err != nil {
		logrus.WithError(err).Fatal("初始化平台超级管理员失败")
	} else if created {
		logrus.Info("已创建平台超级管理员: admin / admin123")
	}

	if err := seed.Run(); err != nil {
		logrus.WithError(err).Fatal("初始化默认数据失败")
	}

	// 加载邮件等通知渠道配置
	emailSettings, err := service.SettingsService{}.GetByCategory("email")
	if err == nil {
		notifier.MustLoadEmailSender(func(key string) string { return emailSettings[key] })
	}

	if *mockData {
		if err := seed.MockOrganizations(); err != nil {
			logrus.WithError(err).Fatal("插入模拟机构数据失败")
		}
	}

	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()
	routes.Register(r)

	logrus.Infof("Base server listening on :%s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		logrus.WithError(err).Fatal("启动服务失败")
	}
}
