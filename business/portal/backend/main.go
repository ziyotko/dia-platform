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
	utils.InitRedisCaptcha()
	utils.InitRedisAnti()

	for _, m := range models.AllModels() {
		utils.DB.AutoMigrate(m)
	}

	// 生产环境使用 release 模式，避免输出敏感调试信息
	gin.SetMode(config.AppConfig.Server.Mode)
	if config.AppConfig.Server.Mode == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(middleware.GinLogger(), gin.Recovery())
	router.SetTrustedProxies([]string{"127.0.0.1"})
	router.Use(middleware.CorsMiddleware())
	router.Use(middleware.SecurityHeaders())

	router.Static(config.AppConfig.Server.UploadDirPrefix+"/uploads", "./uploads")

	routes.SetupRoutes(router)

	addr := fmt.Sprintf("%s:%s", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	utils.Logger.Infof("Server started on %s", addr)
	router.Run(addr)
}
