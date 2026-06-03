package routes

import (
	"github.com/gin-gonic/gin"

	"server/config"
	"server/controllers"
	"server/middleware"
)

func SetupRoutes(router *gin.Engine) {
	authController := controllers.NewAuthController()
	userController := controllers.NewUserController()
	menuController := controllers.NewMenuController()
	roleController := controllers.NewRoleController()
	logController := controllers.NewLogController()
	settingsController := controllers.NewSettingsController()
	uploadController := controllers.NewUploadController()
	templateController := controllers.NewTemplateController()
	apiPrefix := config.AppConfig.Server.ApiPrefix

	public := router.Group(apiPrefix)
	public.Use(middleware.ReplayProtectionMiddleware(), middleware.OperationLog())
	{
		public.GET("/captcha", authController.GetCaptcha)
		public.POST("/login", authController.Login)
		public.GET("/site-info", settingsController.GetPublicSiteInfo)
	}

	protected := router.Group(apiPrefix)
	protected.Use(middleware.ReplayProtectionMiddleware(), middleware.AuthMiddleware(), middleware.OperationLog())
	{
		protected.POST("/logout", authController.Logout)
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

		protected.GET("/settings", settingsController.GetSettings)
		protected.PUT("/settings", settingsController.UpdateSettings)

		protected.POST("/upload", uploadController.UploadFile)

		protected.GET("/templates", templateController.GetTemplates)
		protected.POST("/templates", templateController.CreateTemplate)
		protected.PUT("/templates/:id", templateController.UpdateTemplate)
		protected.DELETE("/templates/:id", templateController.DeleteTemplate)
		protected.PATCH("/templates/:id/status", templateController.UpdateTemplateStatus)
		protected.PUT("/templates/:id/design", templateController.SaveTemplateDesign)
	}
}
