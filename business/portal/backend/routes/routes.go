package routes

import (
	"github.com/gin-gonic/gin"

	"server/config"
	"server/controllers"
	"server/middleware"
)

func SetupRoutes(router *gin.Engine) {
	authController := controllers.NewAuthController()
	workflowController := controllers.NewWorkflowController()
	userController := controllers.NewUserController()
	menuController := controllers.NewMenuController()
	roleController := controllers.NewRoleController()
	workflowRoleController := controllers.NewWorkflowRoleController()
	logController := controllers.NewLogController()
	settingsController := controllers.NewSettingsController()
	uploadController := controllers.NewUploadController()
	templateController := controllers.NewTemplateController()
	pageController := controllers.NewPageController()
	columnController := controllers.NewColumnController()
	categoryController := controllers.NewCategoryController()
	tagController := controllers.NewTagController()
	articleController := controllers.NewArticleController()
	adController := controllers.NewAdController()
	linkController := controllers.NewLinkController()
	deptController := controllers.NewDepartmentController()
	orgController := controllers.NewOrganizationController()
	dashboardController := controllers.NewDashboardController()
	analyticsController := controllers.NewAnalyticsController()
	visitController := controllers.NewVisitController()
	likeController := controllers.NewLikeController()
	shareController := controllers.NewShareController()
	staticLogController := controllers.NewStaticLogController()
	staticPageController := controllers.NewStaticPageController()
	staticMonitorController := controllers.NewStaticMonitorController()
	staticJobController := controllers.NewStaticJobController()
	apiPrefix := config.AppConfig.Server.ApiPrefix

	public := router.Group(apiPrefix)
	public.Use(middleware.ReplayProtectionMiddleware())
	{
		public.GET("/captcha", authController.GetCaptcha)
		public.POST("/login", authController.Login)
		public.GET("/site-info", settingsController.GetPublicSiteInfo)

		//开放文章搜索（无需认证，仅返回已发布文章，不含正文）
		public.GET("/search/articles", articleController.PublicSearchArticles)

		//站点分析接口
		public.POST("/visit", visitController.RecordVisit)
		public.POST("/like", likeController.RecordLike)
		public.POST("/share", shareController.RecordShare)
	}

	ipLimiter := middleware.NewIPLimiter()

	// === 登录用户路由（任意已认证用户可用，需防重放+请求签名）===
	member := router.Group(apiPrefix)
	member.Use(middleware.AuthMiddleware(), middleware.ReplayProtectionMiddleware(), middleware.OperationLog(), ipLimiter.Limit())
	{
		member.POST("/logout", authController.Logout)

		member.GET("/profile", authController.GetProfile)
		member.PUT("/profile", authController.UpdateProfile)
		member.PUT("/profile/password", authController.ChangePassword)

		// 文章管理（作者管理自己的文章，接口内另有归属/管理员校验）
		member.GET("/articles", articleController.GetArticles)
		member.GET("/articles/my-audits", articleController.GetMyAuditArticles)
		member.GET("/articles/author-stats", articleController.GetArticleAuthorStats)
		member.GET("/analytics/article-trend", analyticsController.GetArticleAnalyticsTrend)
		member.GET("/articles/:id", articleController.GetArticleByID)
		member.POST("/articles", articleController.CreateArticle)
		member.PUT("/articles/:id", articleController.UpdateArticle)
		member.DELETE("/articles/:id", articleController.DeleteArticle)
		member.PATCH("/articles/:id/status", articleController.UpdateArticleStatus)
		member.PATCH("/articles/:id/audit", articleController.AuditArticle)
		member.POST("/articles/:id/audit-restart", articleController.RestartArticleAudit)
		member.POST("/articles/:id/audit-withdraw", articleController.WithdrawArticleAudit)
		member.GET("/articles/:id/audit-progress", articleController.GetArticleAuditProgress)
		member.POST("/articles/:id/audit-advance", articleController.AdvanceArticleAudit)
		member.POST("/articles/:id/audit-reject", articleController.RejectArticleAudit)
		member.GET("/articles/:id/audit-history", articleController.GetArticleAuditHistory)
		member.PUT("/articles/:id/columns", articleController.SetArticleColumns)
		member.GET("/articles/column-publishes", articleController.GetArticleColumnPublishes)

		// 内容引用所需的基础数据（只读）
		member.GET("/menus/user", menuController.GetUserMenus)
		member.GET("/settings", settingsController.GetSettings)
		member.GET("/categories", categoryController.GetCategories)
		member.GET("/categories/all", categoryController.GetAllCategories)
		member.GET("/categories/stats", categoryController.GetCategoryArticleStats)
		member.GET("/tags", tagController.GetTags)
		member.GET("/tags/all", tagController.GetAllTags)
		member.GET("/tags/stats", tagController.GetTagArticleStats)
		member.GET("/columns", columnController.GetColumns)
		member.GET("/columns/publishes", columnController.GetColumnPublishes)
		member.GET("/pages", pageController.GetPages)
		member.GET("/static-pages", staticPageController.GetStaticPages)
		member.GET("/workflows", workflowController.GetWorkflows)
		member.GET("/workflows/:id", workflowController.GetWorkflowByID)
		member.GET("/workflows/:id/nodes", workflowController.GetWorkflowNodes)
		member.GET("/workflow-roles", workflowRoleController.GetWorkflowRoles)
		member.GET("/users", userController.GetUsers)
		member.GET("/ads", adController.GetAds)
		member.GET("/ads/:id", adController.GetAdByID)
		member.GET("/links", linkController.GetLinks)
		member.GET("/links/:id", linkController.GetLinkByID)

		// 文件上传（作者上传封面/附件/视频）
		member.POST("/upload", uploadController.UploadFile)
	}

	// === 管理员路由（需认证 + 管理员角色）===
	admin := router.Group(apiPrefix)
	admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware(), middleware.ReplayProtectionMiddleware(), middleware.OperationLog(), ipLimiter.Limit())
	{
		// 用户管理
		admin.POST("/users", userController.CreateUser)
		admin.POST("/users/import", userController.ImportUsers)
		admin.GET("/users/:id", userController.GetUserByID)
		admin.PUT("/users/:id", userController.UpdateUser)
		admin.GET("/users/check-unique", userController.CheckFieldUnique)
		admin.DELETE("/users/:id", userController.DeleteUser)
		admin.PATCH("/users/:id/status", userController.UpdateUserStatus)

		// 菜单管理
		admin.GET("/menus", menuController.GetMenus)
		admin.GET("/menus/tree", menuController.GetMenuTree)
		admin.POST("/menus", menuController.CreateMenu)
		admin.PUT("/menus/:id", menuController.UpdateMenu)
		admin.DELETE("/menus/:id", menuController.DeleteMenu)

		// 角色管理
		admin.GET("/roles", roleController.GetRoles)
		admin.GET("/roles/all", roleController.GetAllRoles)
		admin.GET("/roles/:id/permissions", roleController.GetRolePermissions)
		admin.PUT("/roles/:id/permissions", roleController.UpdateRolePermissions)
		admin.GET("/roles/:id", roleController.GetRoleByID)
		admin.POST("/roles", roleController.CreateRole)
		admin.PUT("/roles/:id", roleController.UpdateRole)
		admin.DELETE("/roles/:id", roleController.DeleteRole)

		// 工作流角色管理
		admin.GET("/workflow-roles/:id", workflowRoleController.GetWorkflowRoleByID)
		admin.POST("/workflow-roles", workflowRoleController.CreateWorkflowRole)
		admin.PUT("/workflow-roles/:id", workflowRoleController.UpdateWorkflowRole)
		admin.DELETE("/workflow-roles/:id", workflowRoleController.DeleteWorkflowRole)
		admin.GET("/workflow-roles/:id/users", workflowRoleController.GetWorkflowRoleUsers)
		admin.PUT("/workflow-roles/:id/users", workflowRoleController.UpdateWorkflowRoleUsers)

		// 工作流管理（写操作）
		admin.POST("/workflows", workflowController.CreateWorkflow)
		admin.PUT("/workflows/:id", workflowController.UpdateWorkflow)
		admin.DELETE("/workflows/:id", workflowController.DeleteWorkflow)
		admin.PUT("/workflows/:id/nodes", workflowController.SaveWorkflowNodes)

		// 日志管理
		admin.GET("/logs", logController.GetLogs)
		admin.DELETE("/logs", logController.ClearLogs)
		admin.GET("/login-logs", logController.GetLoginLogs)
		admin.DELETE("/login-logs", logController.ClearLoginLogs)
		admin.GET("/static-logs", staticLogController.GetLogs)
		admin.DELETE("/static-logs", staticLogController.ClearLogs)
		admin.GET("/static-logs/latest-times", staticLogController.GetLatestTimes)
		admin.GET("/static-monitor", staticMonitorController.GetStaticMonitor)

		// 静态化批量操作任务（代理转发至静态化程序，返回 202 + 任务信息）
		admin.POST("/static/site", staticJobController.SiteStatic)
		admin.POST("/static/pages", staticJobController.PagesStatic)
		admin.POST("/static/lists", staticJobController.ListsStatic)
		admin.POST("/static/articles", staticJobController.ArticlesStatic)
		admin.POST("/static/page", staticJobController.PageStatic)
		admin.POST("/static/list", staticJobController.ListStatic)
		admin.POST("/static/article", staticJobController.ArticleStatic)
		admin.DELETE("/static/article", staticJobController.DeleteArticleStatic)
		admin.GET("/static/jobs/:id", staticJobController.GetJob)

		// 系统设置（写操作）
		admin.PUT("/settings", settingsController.UpdateSettings)

		// 模板管理
		admin.GET("/templates", templateController.GetTemplates)
		admin.POST("/templates", templateController.CreateTemplate)
		admin.PUT("/templates/:id", templateController.UpdateTemplate)
		admin.DELETE("/templates/:id", templateController.DeleteTemplate)
		admin.PATCH("/templates/:id/status", templateController.UpdateTemplateStatus)
		admin.PUT("/templates/:id/design", templateController.SaveTemplateDesign)

		// 页面管理（写操作）
		admin.POST("/pages", pageController.CreatePage)
		admin.PUT("/pages/:id", pageController.UpdatePage)
		admin.DELETE("/pages/:id", pageController.DeletePage)

		// 栏目管理（写操作）
		admin.POST("/columns", columnController.CreateColumn)
		admin.PUT("/columns/:id", columnController.UpdateColumn)
		admin.DELETE("/columns/:id", columnController.DeleteColumn)

		// 分类管理（写操作）
		admin.POST("/categories", categoryController.CreateCategory)
		admin.PUT("/categories/:id", categoryController.UpdateCategory)
		admin.DELETE("/categories/:id", categoryController.DeleteCategory)
		admin.PATCH("/categories/:id/status", categoryController.UpdateCategoryStatus)

		// 标签管理（写操作）
		admin.POST("/tags", tagController.CreateTag)
		admin.PUT("/tags/:id", tagController.UpdateTag)
		admin.DELETE("/tags/:id", tagController.DeleteTag)
		admin.PATCH("/tags/:id/status", tagController.UpdateTagStatus)

		// 广告管理（写操作）
		admin.POST("/ads", adController.CreateAd)
		admin.PUT("/ads/:id", adController.UpdateAd)
		admin.DELETE("/ads/:id", adController.DeleteAd)
		admin.PATCH("/ads/:id/status", adController.UpdateAdStatus)

		// 友情链接管理（写操作）
		admin.POST("/links", linkController.CreateLink)
		admin.PUT("/links/:id", linkController.UpdateLink)
		admin.DELETE("/links/:id", linkController.DeleteLink)
		admin.PATCH("/links/:id/status", linkController.UpdateLinkStatus)

		// 部门管理
		admin.GET("/departments", deptController.GetDepartments)
		admin.GET("/departments/tree", deptController.GetDepartmentTree)
		admin.GET("/departments/:id/users", deptController.GetDepartmentUsers)
		admin.PUT("/departments/:id/users", deptController.AssignDepartmentUsers)
		admin.POST("/departments", deptController.CreateDepartment)
		admin.POST("/departments/import", deptController.ImportDepartments)
		admin.PUT("/departments/:id", deptController.UpdateDepartment)
		admin.DELETE("/departments/:id", deptController.DeleteDepartment)

		// 组织管理
		admin.GET("/organizations", orgController.GetOrganizations)
		admin.GET("/organizations/tree", orgController.GetOrganizationTree)
		admin.GET("/organizations/:id/users", orgController.GetOrganizationUsers)
		admin.PUT("/organizations/:id/users", orgController.AssignOrganizationUsers)
		admin.POST("/organizations", orgController.CreateOrganization)
		admin.PUT("/organizations/:id", orgController.UpdateOrganization)
		admin.DELETE("/organizations/:id", orgController.DeleteOrganization)

		// 仪表盘
		admin.GET("/dashboard/stats", dashboardController.GetStats)
		admin.GET("/dashboard/login-logs", dashboardController.GetLoginLogs)
		admin.GET("/dashboard/visit-trend", dashboardController.GetVisitTrend)
		admin.GET("/dashboard/article-trend", dashboardController.GetArticleTrend)
	}
}
