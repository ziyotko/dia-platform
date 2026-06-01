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

	utils.DB.AutoMigrate(&models.User{}, &models.Menu{})

	initSuperAdmin()
	initMenus()

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

func initMenus() {
	var count int64
	utils.DB.Model(&models.Menu{}).Count(&count)
	if count > 0 {
		return
	}

	menus := []models.Menu{
		{ParentID: 0, Name: "欢迎首页", Path: "/dashboard", Component: "views/dashboard/index.vue", Icon: "HomeFilled", Type: "menu", Sort: 0, Status: 1},
		{ParentID: 0, Name: "系统管理", Path: "/system", Component: "", Icon: "Tools", Type: "directory", Sort: 1, Status: 1},
		{ParentID: 0, Name: "内容管理", Path: "/content", Component: "", Icon: "Document", Type: "directory", Sort: 2, Status: 1},
		{ParentID: 0, Name: "个人中心", Path: "/personal", Component: "", Icon: "User", Type: "directory", Sort: 3, Status: 1},
	}

	for i := range menus {
		if err := utils.DB.Create(&menus[i]).Error; err != nil {
			utils.Logger.Errorf("Failed to create menu: %v", err)
		}
	}

	var systemID, contentID, personalID uint
	utils.DB.Model(&models.Menu{}).Where("name = ?", "系统管理").Select("id").Scan(&systemID)
	utils.DB.Model(&models.Menu{}).Where("name = ?", "内容管理").Select("id").Scan(&contentID)
	utils.DB.Model(&models.Menu{}).Where("name = ?", "个人中心").Select("id").Scan(&personalID)

	subMenus := []models.Menu{
		{ParentID: systemID, Name: "用户管理", Path: "/users", Component: "views/system/users.vue", Icon: "UserFilled", Type: "menu", Sort: 1, Status: 1},
		{ParentID: systemID, Name: "角色管理", Path: "/roles", Component: "views/system/roles.vue", Icon: "Avatar", Type: "menu", Sort: 2, Status: 1},
		{ParentID: systemID, Name: "菜单管理", Path: "/menus", Component: "views/system/menus.vue", Icon: "Menu", Type: "menu", Sort: 3, Status: 1},
		{ParentID: systemID, Name: "操作日志", Path: "/logs", Component: "views/system/logs.vue", Icon: "List", Type: "menu", Sort: 4, Status: 1},
		{ParentID: contentID, Name: "文章管理", Path: "/content/article", Component: "views/content/article.vue", Icon: "Document", Type: "menu", Sort: 1, Status: 1},
		{ParentID: contentID, Name: "分类管理", Path: "/content/category", Component: "views/content/category.vue", Icon: "Folder", Type: "menu", Sort: 2, Status: 1},
		{ParentID: contentID, Name: "标签管理", Path: "/content/tag", Component: "views/content/tag.vue", Icon: "PriceTag", Type: "menu", Sort: 3, Status: 1},
		{ParentID: contentID, Name: "评论管理", Path: "/content/comment", Component: "views/content/comment.vue", Icon: "ChatDotSquare", Type: "menu", Sort: 4, Status: 1},
		{ParentID: contentID, Name: "广告管理", Path: "/content/ad", Component: "views/content/ad.vue", Icon: "Promotion", Type: "menu", Sort: 5, Status: 1},
		{ParentID: contentID, Name: "友链管理", Path: "/content/link", Component: "views/content/link.vue", Icon: "Link", Type: "menu", Sort: 6, Status: 1},
		{ParentID: personalID, Name: "个人信息", Path: "/profile", Component: "views/profile/index.vue", Icon: "Document", Type: "menu", Sort: 1, Status: 1},
		{ParentID: personalID, Name: "系统设置", Path: "/settings", Component: "views/settings/index.vue", Icon: "Setting", Type: "menu", Sort: 2, Status: 1},
	}

	for i := range subMenus {
		if err := utils.DB.Create(&subMenus[i]).Error; err != nil {
			utils.Logger.Errorf("Failed to create sub menu: %v", err)
		}
	}

	utils.Logger.Info("Menus initialized successfully")
}
