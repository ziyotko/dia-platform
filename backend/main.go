package main

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"server/config"
	"server/middleware"
	"server/models"
	"server/routes"
	"server/utils"
)

func main() {
	config.InitConfig()
	utils.InitLogger()
	utils.InitDB()
	utils.InitRedis()

	utils.DB.AutoMigrate(&models.User{})

	initSuperAdmin()

	router := gin.New()
	router.Use(middleware.GinLogger(), gin.Recovery())
	router.SetTrustedProxies([]string{"127.0.0.1"})
	router.Use(middleware.CorsMiddleware())

	routes.SetupRoutes(router)

	addr := fmt.Sprintf("%s:%s", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	utils.Logger.Infof("Server started on %s", addr)
	router.Run(addr)
}

func initSuperAdmin() {
	var count int64
	utils.DB.Model(&models.User{}).Where("account = ?", "superadmin").Count(&count)
	if count > 0 {
		return
	}

	admin := &models.User{
		Account:  "superadmin",
		Username: "超级管理员",
		Password: "123456",
		Email:    "superadmin@example.com",
		Mobile:   "13800138000",
		Status:   1,
	}

	if err := utils.DB.Create(admin).Error; err != nil {
		utils.Logger.Errorf("Failed to create super admin: %v", err)
	} else {
		utils.Logger.Info("Super admin created successfully: superadmin / 123456")
	}
}