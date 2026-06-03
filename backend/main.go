package main

import (
	"fmt"
	"strconv"
	"strings"

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

	utils.DB.AutoMigrate(&models.User{}, &models.Menu{}, &models.Role{}, &models.OperationLog{}, &models.Settings{}, &models.Template{}, &models.Page{})

	initSuperAdmin()
	initMenus()
	initRoles()

	router := gin.New()
	router.Use(middleware.GinLogger(), gin.Recovery())
	router.SetTrustedProxies([]string{"127.0.0.1"})
	router.Use(middleware.CorsMiddleware())

	router.Static("/uploads", "./uploads")

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

func initRoles() {
	var count int64
	utils.DB.Model(&models.Role{}).Count(&count)
	if count > 0 {
		return
	}

	var menus []models.Menu
	if err := utils.DB.Find(&menus).Error; err != nil {
		utils.Logger.Errorf("Failed to get menus for role init: %v", err)
		return
	}

	menuIDMap := make(map[string]uint)
	var allMenuIDs []string
	for _, m := range menus {
		menuIDMap[m.Name] = m.ID
		allMenuIDs = append(allMenuIDs, strconv.FormatUint(uint64(m.ID), 10))
	}

	joinIDs := func(names ...string) string {
		ids := []string{}
		for _, name := range names {
			if id, ok := menuIDMap[name]; ok {
				ids = append(ids, strconv.FormatUint(uint64(id), 10))
			}
		}
		return strings.Join(ids, ",")
	}

	roles := []models.Role{
		{
			Name:        "系统管理员",
			Code:        "super_admin",
			Description: "拥有全站配置、用户管理、角色权限分配、安全设置等最高权限，不可被其他角色替代",
			Status:      1,
			Permissions: strings.Join(allMenuIDs, ","),
		},
		{
			Name:        "内容编辑人员",
			Code:        "operator",
			Description: "负责发布、编辑、删除内容（文章、页面、媒体等），通常无权管理用户或系统设置",
			Status:      1,
			Permissions: joinIDs("欢迎首页", "内容管理", "文章管理", "分类管理", "标签管理", "评论管理", "广告管理", "友链管理", "个人中心", "个人信息", "系统设置"),
		},
		{
			Name:        "审批人",
			Code:        "approver",
			Description: "专门审核待发布内容或敏感操作（如删除、置顶），确保合规性，权限通常仅限审批流相关功能",
			Status:      1,
			Permissions: joinIDs("欢迎首页", "文章管理", "评论管理", "个人中心", "个人信息"),
		},
		{
			Name:        "投稿人",
			Code:        "contributor",
			Description: "可撰写或上传内容，但需经审核才能发布，不能直接发布或修改他人内容",
			Status:      1,
			Permissions: joinIDs("欢迎首页", "文章管理", "个人中心", "个人信息"),
		},
		{
			Name:        "访客",
			Code:        "visitor",
			Description: "仅查看后台数据或报表（如数据分析岗），无编辑权限，多见于内部协作型门户",
			Status:      1,
			Permissions: joinIDs("欢迎首页", "个人中心", "个人信息"),
		},
	}

	for i := range roles {
		if err := utils.DB.Create(&roles[i]).Error; err != nil {
			utils.Logger.Errorf("Failed to create role %s: %v", roles[i].Code, err)
		}
	}

	utils.Logger.Info("Roles initialized successfully")
}
