package main

import (
	"member/config"
	"member/internal/middleware"
	"member/internal/models"
	"member/internal/routes"
	"member/internal/seed"
	"member/pkg/db"
	"member/pkg/redis"
	"member/pkg/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load config
	config.Load("config.yaml")

	// Init logger
	utils.InitLogger()

	// Init database
	db.Init(&config.Cfg.MySQL)

	// Init Redis
	redis.Init(&config.Cfg.Redis)

	// Captcha is initialized on-demand via redis store

	// Init IP limiter
	middleware.InitIPLimiter(config.Cfg.Server.MaxConcurrentIPs)

	// Auto migrate all models
	if err := db.DB.AutoMigrate(
		&models.Member{},
		&models.Application{},
		&models.Certificate{},
		&models.FeeRecord{},
		&models.MemberMessage{},
		&models.Organization{},
		&models.MemberOrganization{},
		&models.Article{},
		&models.ArticleCategory{},
		&models.Announcement{},
		&models.SystemConfig{},
		&models.PasswordReset{},
		&models.MemberLevel{},
		&models.MemberOrgLevel{},
		&models.MemberFeeStandard{},
		&models.MemberCertificateTemplate{},
		&models.MemberLevelChange{},
		&models.ProfileChange{},
		&models.OperationLog{},
	); err != nil {
		utils.Logger.Fatalf("AutoMigrate failed: %v", err)
	}

	// 软删 -> 硬删 收尾：物理清除历史软删记录并删除已无用的 deleted_at 列
	db.PurgeSoftDeleted()

	// Seed default data
	seed.Run()

	// Setup Gin
	gin.SetMode(config.Cfg.Server.Mode)
	r := gin.New()

	// Global middleware
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(middleware.IPLimit())
	r.Use(gin.Recovery())
	r.Use(middleware.SecureUploads())
	r.SetTrustedProxies(config.Cfg.Server.TrustedProxies)

	// Serve uploaded files
	r.Static(config.Cfg.Server.UploadDirPrefix+"/uploads", "./uploads")

	// Register routes
	routes.Register(r)

	// Start server
	addr := config.Cfg.Server.Host + ":" + itoa(config.Cfg.Server.Port)
	utils.Logger.Infof("Member server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		utils.Logger.Fatalf("Server failed: %v", err)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
