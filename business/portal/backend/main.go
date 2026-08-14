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

	// 初始化系统默认角色（超级管理员/普通管理员/内容审核/内容作者），缺失时自动创建
	models.SeedDefaultRoles()
	// 初始化系统默认用户（与默认角色一一对应，初始密码 1qaz@WSX），缺失时自动创建
	models.SeedDefaultUsers()
	// 初始化系统默认菜单（管理首页/内容管理/数据统计/基础配置/系统配置），缺失时自动创建
	models.SeedDefaultMenus()

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
	fmt.Printf("Server started on %s\n", addr)
	router.Run(addr)
}
