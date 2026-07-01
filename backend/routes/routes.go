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
	visitController := controllers.NewVisitController()
	staticLogController := controllers.NewStaticLogController()
	staticPageController := controllers.NewStaticPageController()
	apiPrefix := config.AppConfig.Server.ApiPrefix

	public := router.Group(apiPrefix)
	public.Use(middleware.ReplayProtectionMiddleware())
	{
		public.GET("/captcha", authController.GetCaptcha)
		public.POST("/login", authController.Login)
		public.GET("/site-info", settingsController.GetPublicSiteInfo)
		public.POST("/visit", visitController.RecordVisit)
	}

	protected := router.Group(apiPrefix)
	protected.Use(middleware.AuthMiddleware(), middleware.ReplayProtectionMiddleware(), middleware.OperationLog())
	{
		protected.POST("/logout", authController.Logout)

		protected.GET("/workflows", workflowController.GetWorkflows)
		protected.POST("/workflows", workflowController.CreateWorkflow)
		protected.GET("/workflows/:id", workflowController.GetWorkflowByID)
		protected.PUT("/workflows/:id", workflowController.UpdateWorkflow)
		protected.DELETE("/workflows/:id", workflowController.DeleteWorkflow)
		protected.PUT("/workflows/:id/nodes", workflowController.SaveWorkflowNodes)
		protected.GET("/workflows/:id/nodes", workflowController.GetWorkflowNodes)

		protected.GET("/profile", authController.GetProfile)
		protected.PUT("/profile", authController.UpdateProfile)
		protected.PUT("/profile/password", authController.ChangePassword)

		protected.GET("/users", userController.GetUsers)
		protected.POST("/users", userController.CreateUser)
		protected.GET("/users/:id", userController.GetUserByID)
		protected.PUT("/users/:id", userController.UpdateUser)
		protected.DELETE("/users/:id", userController.DeleteUser)
		protected.PATCH("/users/:id/status", userController.UpdateUserStatus)

		protected.GET("/menus", menuController.GetMenus)
		protected.GET("/menus/tree", menuController.GetMenuTree)
		protected.GET("/menus/user", menuController.GetUserMenus)
		protected.POST("/menus", menuController.CreateMenu)
		protected.PUT("/menus/:id", menuController.UpdateMenu)
		protected.DELETE("/menus/:id", menuController.DeleteMenu)

		protected.GET("/roles", roleController.GetRoles)
		protected.GET("/roles/all", roleController.GetAllRoles)
		protected.GET("/roles/:id/permissions", roleController.GetRolePermissions)
		protected.PUT("/roles/:id/permissions", roleController.UpdateRolePermissions)
		protected.GET("/roles/:id", roleController.GetRoleByID)
		protected.POST("/roles", roleController.CreateRole)
		protected.PUT("/roles/:id", roleController.UpdateRole)
		protected.DELETE("/roles/:id", roleController.DeleteRole)

		protected.GET("/logs", logController.GetLogs)
		protected.DELETE("/logs", logController.ClearLogs)
		protected.GET("/login-logs", logController.GetLoginLogs)

		protected.GET("/static-logs", staticLogController.GetLogs)
		protected.DELETE("/static-logs", staticLogController.ClearLogs)

		protected.GET("/settings", settingsController.GetSettings)
		protected.PUT("/settings", settingsController.UpdateSettings)

		protected.POST("/upload", uploadController.UploadFile)

		protected.GET("/templates", templateController.GetTemplates)
		protected.POST("/templates", templateController.CreateTemplate)
		protected.PUT("/templates/:id", templateController.UpdateTemplate)
		protected.DELETE("/templates/:id", templateController.DeleteTemplate)
		protected.PATCH("/templates/:id/status", templateController.UpdateTemplateStatus)
		protected.PUT("/templates/:id/design", templateController.SaveTemplateDesign)

		protected.GET("/pages", pageController.GetPages)
		protected.POST("/pages", pageController.CreatePage)
		protected.PUT("/pages/:id", pageController.UpdatePage)
		protected.DELETE("/pages/:id", pageController.DeletePage)

		protected.GET("/static-pages", staticPageController.GetStaticPages)

		protected.GET("/columns", columnController.GetColumns)
		protected.POST("/columns", columnController.CreateColumn)
		protected.PUT("/columns/:id", columnController.UpdateColumn)
		protected.DELETE("/columns/:id", columnController.DeleteColumn)

		protected.GET("/categories", categoryController.GetCategories)
		protected.GET("/categories/all", categoryController.GetAllCategories)
		protected.GET("/categories/stats", categoryController.GetCategoryArticleStats)
		protected.POST("/categories", categoryController.CreateCategory)
		protected.PUT("/categories/:id", categoryController.UpdateCategory)
		protected.DELETE("/categories/:id", categoryController.DeleteCategory)
		protected.PATCH("/categories/:id/status", categoryController.UpdateCategoryStatus)

		protected.GET("/tags", tagController.GetTags)
		protected.GET("/tags/all", tagController.GetAllTags)
		protected.GET("/tags/stats", tagController.GetTagArticleStats)
		protected.POST("/tags", tagController.CreateTag)
		protected.PUT("/tags/:id", tagController.UpdateTag)
		protected.DELETE("/tags/:id", tagController.DeleteTag)
		protected.PATCH("/tags/:id/status", tagController.UpdateTagStatus)

		protected.GET("/articles", articleController.GetArticles)
		protected.GET("/articles/my-audits", articleController.GetMyAuditArticles)
		protected.GET("/articles/:id", articleController.GetArticleByID)
		protected.POST("/articles", articleController.CreateArticle)
		protected.PUT("/articles/:id", articleController.UpdateArticle)
		protected.DELETE("/articles/:id", articleController.DeleteArticle)
		protected.PATCH("/articles/:id/status", articleController.UpdateArticleStatus)
		protected.PATCH("/articles/:id/audit", articleController.AuditArticle)
		protected.POST("/articles/:id/audit-restart", articleController.RestartArticleAudit)
		protected.POST("/articles/:id/audit-withdraw", articleController.WithdrawArticleAudit)
		protected.GET("/articles/:id/audit-progress", articleController.GetArticleAuditProgress)
		protected.POST("/articles/:id/audit-advance", articleController.AdvanceArticleAudit)
		protected.POST("/articles/:id/audit-reject", articleController.RejectArticleAudit)
		protected.GET("/articles/:id/audit-history", articleController.GetArticleAuditHistory)
		protected.PUT("/articles/:id/columns", articleController.SetArticleColumns)
		protected.GET("/articles/column-publishes", articleController.GetArticleColumnPublishes)

		protected.GET("/ads", adController.GetAds)
		protected.GET("/ads/:id", adController.GetAdByID)
		protected.POST("/ads", adController.CreateAd)
		protected.PUT("/ads/:id", adController.UpdateAd)
		protected.DELETE("/ads/:id", adController.DeleteAd)
		protected.PATCH("/ads/:id/status", adController.UpdateAdStatus)

		protected.GET("/links", linkController.GetLinks)
		protected.GET("/links/:id", linkController.GetLinkByID)
		protected.POST("/links", linkController.CreateLink)
		protected.PUT("/links/:id", linkController.UpdateLink)
		protected.DELETE("/links/:id", linkController.DeleteLink)
		protected.PATCH("/links/:id/status", linkController.UpdateLinkStatus)

		protected.GET("/departments", deptController.GetDepartments)
		protected.GET("/departments/tree", deptController.GetDepartmentTree)
		protected.GET("/departments/:id/users", deptController.GetDepartmentUsers)
		protected.PUT("/departments/:id/users", deptController.AssignDepartmentUsers)
		protected.POST("/departments", deptController.CreateDepartment)
		protected.PUT("/departments/:id", deptController.UpdateDepartment)
		protected.DELETE("/departments/:id", deptController.DeleteDepartment)

		protected.GET("/organizations", orgController.GetOrganizations)
		protected.GET("/organizations/tree", orgController.GetOrganizationTree)
		protected.GET("/organizations/:id/users", orgController.GetOrganizationUsers)
		protected.PUT("/organizations/:id/users", orgController.AssignOrganizationUsers)
		protected.POST("/organizations", orgController.CreateOrganization)
		protected.PUT("/organizations/:id", orgController.UpdateOrganization)
		protected.DELETE("/organizations/:id", orgController.DeleteOrganization)

		protected.GET("/dashboard/stats", dashboardController.GetStats)
		protected.GET("/dashboard/login-logs", dashboardController.GetLoginLogs)
		protected.GET("/dashboard/visit-trend", dashboardController.GetVisitTrend)
	}
}
