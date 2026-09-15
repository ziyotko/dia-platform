package routes

import (
	"time"

	"base/config"
	"base/internal/adapter"
	"base/internal/controllers"
	"base/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Register(r *gin.Engine) {
	r.Use(middleware.CORS())
	// 可信反向代理：仅当请求来自这些地址时才信任 X-Forwarded-For / X-Real-IP，
	// 否则一律使用 RemoteAddr，防伪造头绕过按 IP 的限流
	if err := r.SetTrustedProxies(config.Cfg.Server.TrustedProxies); err != nil {
		logrus.WithError(err).Warn("设置可信代理失败，将只使用 RemoteAddr 作为客户端 IP")
	}

	srv := config.Cfg.Server
	api := r.Group("/base/api/v1")

	// 文件控制器持有存储依赖，需实例化后复用（不能用 &FileController{} 零值）
	fileCtl := controllers.NewFileController()

	// 公开接口（均按真实客户端 IP 限流，参数见 config.yaml 的 server.*_rate_limit）
	auth := api.Group("/auth")
	{
		auth.POST("/login", middleware.RateLimitMiddleware(
			srv.LoginRateLimit,
			time.Duration(srv.LoginRateWindowSecs)*time.Second,
		), (&controllers.AuthController{}).Login)
		auth.POST("/init", middleware.RateLimitMiddleware(
			srv.InitRateLimit,
			time.Duration(srv.InitRateWindowSecs)*time.Second,
		), (&controllers.AuthController{}).InitAdmin)
		auth.GET("/captcha", middleware.RateLimitMiddleware(
			srv.CaptchaRateLimit,
			time.Duration(srv.CaptchaRateWindowSecs)*time.Second,
		), (&controllers.AuthController{}).Captcha)
	}

	// 公开站点信息（登录页读取验证码开关等）：与验证码接口同量级，按相同限流参数保护
	api.GET("/site-info", middleware.RateLimitMiddleware(
		srv.CaptchaRateLimit,
		time.Duration(srv.CaptchaRateWindowSecs)*time.Second,
	), (&controllers.SettingsController{}).SiteInfo)

	// 文件公开访问（key 含日期目录，使用通配路由）
	api.GET("/files/*key", fileCtl.Serve)

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
			msgs.POST("/read-all", (&controllers.MessageController{}).MarkAllRead)
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
			files.POST("/upload", fileCtl.Upload)
			files.GET("", fileCtl.List)
			files.DELETE("/:id", fileCtl.Delete)
		}
	}

	// 子应用统一代理入口
	r.Any("/base/api/v1/app/:appCode/*path", middleware.JWTAuth(), adapter.AdapterController)
}
