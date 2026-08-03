package main

import (
	"conference/config"
	"conference/internal/middleware"
	"conference/internal/models"
	"conference/internal/routes"
	"conference/internal/seed"
	"conference/pkg/captcha"
	"conference/pkg/db"
	"conference/pkg/redis"
	"conference/pkg/utils"

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
		&models.Meeting{},
		&models.Agenda{},
		&models.Guest{},
		&models.Registration{},
		&models.SignIn{},
		&models.Vote{},
		&models.VoteOption{},
		&models.VoteRecord{},
		&models.Order{},
		&models.Refund{},
		&models.Invoice{},
		&models.LiveConfig{},
		&models.VodConfig{},
		&models.LiveMessage{},
		&models.ViewingLog{},
		&models.Survey{},
		&models.SurveyQuestion{},
		&models.SurveyOption{},
		&models.SurveyAnswer{},
		&models.CreditConfig{},
		&models.CreditRecord{},
		&models.Archive{},
		&models.ArchiveItem{},
		&models.Notification{},
		&models.NotificationRead{},
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

	// 10. Register routes
	routes.Register(r)

	// 11. Start server
	utils.Logger.Info("Conference server starting on port " + itoa(config.Cfg.Server.Port))
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
