package main

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"server/config"
	"server/middleware"
	"server/models"
	"server/routes"
	"server/services"
	"server/utils"
)

func main() {
	config.InitConfig()
	utils.InitLogger()
	utils.InitDB()
	utils.InitRedisCaptcha()
	utils.InitRedisCache()
	utils.InitRedisAnti()

	for _, m := range models.AllModels() {
		utils.DB.AutoMigrate(m)
	}
	// FULLTEXT 索引不在 AutoMigrate 能力范围内（文章搜索的 MATCH ... AGAINST 依赖它），启动时幂等补齐
	models.EnsureFulltextIndexes()

	// 初始化系统默认角色（管理员/内容审核/内容作者），缺失时自动创建
	models.SeedDefaultRoles()
	// 初始化系统默认用户（与默认角色一一对应，初始密码 1qaz@WSX），缺失时自动创建
	models.SeedDefaultUsers()
	// 初始化系统默认菜单（管理首页/内容管理/数据统计/基础配置/系统配置），缺失时自动创建
	models.SeedDefaultMenus()
	// 初始化内置角色（内容审核/内容作者）的默认菜单权限，仅在角色尚无权限配置时写入
	models.SeedDefaultRolePermissions()

	// 启动时加载静态化参数（静态化输出路径/访问地址/访问令牌名/首页整体变灰）并写入缓存
	settingsService := &services.SettingsService{}
	if err := settingsService.LoadStaticParamsToCache(); err != nil {
		utils.Logger.Warnf("加载静态化参数到缓存失败: %s", err)
	} else {
		utils.Logger.Infof("静态化参数已加载到缓存 (key: %s)", services.StaticParamsCacheKey)
	}

	// 生产环境使用 release 模式，避免输出敏感调试信息
	gin.SetMode(config.AppConfig.Server.Mode)
	if config.AppConfig.Server.Mode == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(middleware.GinLogger(), gin.Recovery())

	// 可信反向代理读取自配置；若为空则回退为仅信任本机回环，避免误信任任意代理头
	trustedProxies := config.AppConfig.Server.TrustedProxies
	if len(trustedProxies) == 0 {
		trustedProxies = []string{"127.0.0.1"}
	}
	if err := router.SetTrustedProxies(trustedProxies); err != nil {
		utils.Logger.Warnf("设置可信代理失败，回退为仅信任回环地址: %s", err)
		_ = router.SetTrustedProxies([]string{"127.0.0.1"})
	}

	router.Use(middleware.CorsMiddleware())
	router.Use(middleware.SecurityHeaders())

	router.Static(config.AppConfig.Server.UploadDirPrefix+"/uploads", "./uploads")

	routes.SetupRoutes(router)

	addr := fmt.Sprintf("%s:%s", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	utils.Logger.Infof("Server started on %s", addr)
	router.Run(addr)
}
