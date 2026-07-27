package routes

import (
	"base/internal/adapter"
	"base/internal/controllers"
	"base/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	r.Use(middleware.CORS())
	r.SetTrustedProxies([]string{"127.0.0.1"})

	api := r.Group("/base/api/v1")

	// 公开接口
	auth := api.Group("/auth")
	{
		auth.POST("/login", (&controllers.AuthController{}).Login)
		auth.POST("/init", (&controllers.AuthController{}).InitAdmin)
		auth.GET("/captcha", (&controllers.AuthController{}).Captcha)
	}

	// 需要登录
	// 文件公开访问
	api.GET("/files/:key", (&controllers.FileController{}).Serve)

	authorized := api.Group("", middleware.JWTAuth())
	authorized.Use(middleware.OperationLog())
	authorized.Use(middleware.PermissionAuth())
	{
		authorized.GET("/auth/info", (&controllers.AuthController{}).Info)
		authorized.GET("/auth/menus", (&controllers.AuthController{}).Menus)
		authorized.GET("/auth/permissions", (&controllers.AuthController{}).Permissions)
		authorized.POST("/auth/change-password", (&controllers.AuthController{}).ChangePassword)

		authorized.GET("/dashboard/stats", (&controllers.DashboardController{}).Stats)

		tenants := authorized.Group("/tenants", middleware.SuperAdminOnly())
		{
			tenants.POST("", (&controllers.TenantController{}).Create)
			tenants.PUT("/:id", (&controllers.TenantController{}).Update)
			tenants.DELETE("/:id", (&controllers.TenantController{}).Delete)
			tenants.GET("/:id", (&controllers.TenantController{}).Get)
			tenants.GET("", (&controllers.TenantController{}).List)
		}

		settings := authorized.Group("/settings")
		{
			settings.GET("", (&controllers.SettingsController{}).Get)
			settings.PUT("", (&controllers.SettingsController{}).Save)
			settings.POST("/email/test", (&controllers.SettingsController{}).TestEmail)
		}

		dicts := authorized.Group("/dicts")
		{
			dicts.POST("", (&controllers.DictController{}).Create)
			dicts.PUT("/:id", (&controllers.DictController{}).Update)
			dicts.DELETE("/:id", (&controllers.DictController{}).Delete)
			dicts.GET("/:id", (&controllers.DictController{}).Get)
			dicts.GET("", (&controllers.DictController{}).List)
			dicts.GET("/code/:code", (&controllers.DictController{}).GetByCode)
			dicts.POST("/:id/items", (&controllers.DictController{}).SaveItems)
		}

		apps := authorized.Group("/apps")
		{
			apps.POST("", (&controllers.AppController{}).Create)
			apps.PUT("/:id", (&controllers.AppController{}).Update)
			apps.DELETE("/:id", (&controllers.AppController{}).Delete)
			apps.GET("/:id", (&controllers.AppController{}).Get)
			apps.GET("", (&controllers.AppController{}).List)
		}

		instances := authorized.Group("/app-instances")
		{
			instances.POST("", (&controllers.AppInstanceController{}).Create)
			instances.PUT("/:id", (&controllers.AppInstanceController{}).Update)
			instances.DELETE("/:id", (&controllers.AppInstanceController{}).Delete)
			instances.GET("", (&controllers.AppInstanceController{}).List)
			instances.GET("/my", (&controllers.AppInstanceController{}).MyApps)
		}

		users := authorized.Group("/users")
		{
			users.POST("", (&controllers.UserController{}).Create)
			users.PUT("/:id", (&controllers.UserController{}).Update)
			users.DELETE("/:id", (&controllers.UserController{}).Delete)
			users.GET("/:id", (&controllers.UserController{}).Get)
			users.GET("", (&controllers.UserController{}).List)
			users.POST("/:id/roles", (&controllers.UserController{}).AssignRoles)
			users.POST("/:id/reset-password", (&controllers.UserController{}).ResetPassword)
		}

		roles := authorized.Group("/roles")
		{
			roles.POST("", (&controllers.RoleController{}).Create)
			roles.PUT("/:id", (&controllers.RoleController{}).Update)
			roles.DELETE("/:id", (&controllers.RoleController{}).Delete)
			roles.GET("/:id", (&controllers.RoleController{}).Get)
			roles.GET("", (&controllers.RoleController{}).List)
			roles.POST("/:id/menus", (&controllers.RoleController{}).AssignMenus)
			roles.POST("/:id/permissions", (&controllers.RoleController{}).AssignPermissions)
		}

		menus := authorized.Group("/menus")
		{
			menus.POST("", (&controllers.MenuController{}).Create)
			menus.PUT("/:id", (&controllers.MenuController{}).Update)
			menus.DELETE("/:id", (&controllers.MenuController{}).Delete)
			menus.GET("/tree", (&controllers.MenuController{}).Tree)
		}

		perms := authorized.Group("/permissions")
		{
			perms.POST("", (&controllers.PermissionController{}).Create)
			perms.PUT("/:id", (&controllers.PermissionController{}).Update)
			perms.DELETE("/:id", (&controllers.PermissionController{}).Delete)
			perms.GET("/tree", (&controllers.PermissionController{}).Tree)
		}

		orgs := authorized.Group("/organizations")
		{
			orgs.POST("", (&controllers.OrganizationController{}).Create)
			orgs.PUT("/:id", (&controllers.OrganizationController{}).Update)
			orgs.DELETE("/:id", (&controllers.OrganizationController{}).Delete)
			orgs.GET("/:id", (&controllers.OrganizationController{}).Get)
			orgs.GET("/tree", (&controllers.OrganizationController{}).Tree)
		}

		msgs := authorized.Group("/messages")
		{
			msgs.GET("", (&controllers.MessageController{}).List)
			msgs.GET("/unread-count", (&controllers.MessageController{}).UnreadCount)
			msgs.GET("/:id", (&controllers.MessageController{}).Get)
			msgs.POST("", (&controllers.MessageController{}).Create)
			msgs.POST("/send", (&controllers.MessageController{}).Send)
			msgs.POST("/:id/read", (&controllers.MessageController{}).MarkRead)
			msgs.DELETE("/:id", (&controllers.MessageController{}).Delete)
		}

		templates := authorized.Group("/message-templates")
		{
			templates.POST("", (&controllers.MessageTemplateController{}).Create)
			templates.PUT("/:id", (&controllers.MessageTemplateController{}).Update)
			templates.DELETE("/:id", (&controllers.MessageTemplateController{}).Delete)
			templates.GET("/:id", (&controllers.MessageTemplateController{}).Get)
			templates.GET("", (&controllers.MessageTemplateController{}).List)
		}

		logs := authorized.Group("/operation-logs")
		{
			logs.GET("", (&controllers.OperationLogController{}).List)
			logs.POST("/delete", (&controllers.OperationLogController{}).Delete)
			logs.POST("/clear", (&controllers.OperationLogController{}).Clear)
			logs.GET("/export", (&controllers.OperationLogController{}).Export)
		}

		loginLogs := authorized.Group("/login-logs")
		{
			loginLogs.GET("", (&controllers.LoginLogController{}).List)
			loginLogs.POST("/delete", (&controllers.LoginLogController{}).Delete)
			loginLogs.POST("/clear", (&controllers.LoginLogController{}).Clear)
			loginLogs.GET("/export", (&controllers.LoginLogController{}).Export)
		}

		files := authorized.Group("/files")
		{
			files.POST("/upload", (&controllers.FileController{}).Upload)
			files.GET("", (&controllers.FileController{}).List)
			files.DELETE("/:id", (&controllers.FileController{}).Delete)
		}
	}

	// 子应用统一代理入口
	r.Any("/base/api/v1/app/:appCode/*path", middleware.JWTAuth(), adapter.AdapterController)
}
