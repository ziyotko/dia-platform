package main

import (
	"application/config"
	"application/internal/middleware"
	"application/internal/models"
	"application/internal/routes"
	"application/internal/seed"
	"application/pkg/captcha"
	"application/pkg/db"
	"application/pkg/redis"
	"application/pkg/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load config
	config.Load("config.yaml")

	// 2. Init logger
	utils.InitLogger()

	// 3. Init DB
	db.Init(&config.Cfg.MySQL)

	// 4. Init Redis
	redis.Init(&config.Cfg.Redis)

	// 5. Init captcha store
	captcha.Init()

	// 6. Init IP limiter
	middleware.InitIPLimiter(config.Cfg.Server.MaxConcurrentIPs)

	// 7. AutoMigrate
	db.DB.AutoMigrate(
		&models.User{},
		&models.Admin{},
		&models.Role{},
		&models.Expert{},
		&models.ProjectCategory{},
		&models.ProjectBatch{},
		&models.Application{},
		&models.ApplicationMaterial{},
		&models.ReviewAssignment{},
		&models.Certificate{},
		&models.Announcement{},
		&models.Notification{},
		&models.AuditLog{},
		&seed.SystemConfig{},
	)

	// 8. Seed default data
	seed.Run()

	// 9. Setup Gin
	gin.SetMode(config.Cfg.Server.Mode)
	r := gin.New()
	r.Use(middleware.CORS(), middleware.Logger(), middleware.IPLimit(), gin.Recovery())
	r.Static("/uploads", "./uploads")
	r.MaxMultipartMemory = 64 << 20 // 64MB

	// 10. Register routes
	routes.Register(r)

	// 11. Start server
	utils.Logger.Info("Application server starting on port " + itoa(config.Cfg.Server.Port))
	r.Run("0.0.0.0:" + itoa(config.Cfg.Server.Port))
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
