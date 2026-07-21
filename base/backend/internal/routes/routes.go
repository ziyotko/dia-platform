package routes

import (
	"base/internal/adapter"
	"base/internal/controllers"
	"base/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	r.Use(middleware.CORS())

	api := r.Group("/base/api/v1")

	// 公开接口
	auth := api.Group("/auth")
	{
		auth.POST("/login", (&controllers.AuthController{}).Login)
		auth.POST("/init", (&controllers.AuthController{}).InitAdmin)
	}

	// 需要登录
	authorized := api.Group("", middleware.JWTAuth())
	authorized.Use(middleware.OperationLog())
	{
		authorized.GET("/auth/info", (&controllers.AuthController{}).Info)
		authorized.GET("/auth/menus", (&controllers.AuthController{}).Menus)
		authorized.GET("/auth/permissions", (&controllers.AuthController{}).Permissions)
		authorized.POST("/auth/change-password", (&controllers.AuthController{}).ChangePassword)

		tenants := authorized.Group("/tenants", middleware.SuperAdminOnly())
		{
			tenants.POST("", (&controllers.TenantController{}).Create)
			tenants.PUT("/:id", (&controllers.TenantController{}).Update)
			tenants.DELETE("/:id", (&controllers.TenantController{}).Delete)
			tenants.GET("/:id", (&controllers.TenantController{}).Get)
			tenants.GET("", (&controllers.TenantController{}).List)
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
	}

	// 子应用统一代理入口
	r.Any("/base/api/v1/app/:appCode/*path", middleware.JWTAuth(), adapter.AdapterController)
}
