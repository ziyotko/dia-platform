package routes

import (
	"github.com/gin-gonic/gin"

	"server/config"
	"server/controllers"
	"server/middleware"
)

func SetupRoutes(router *gin.Engine) {
	authController := controllers.NewAuthController()
	apiPrefix := config.AppConfig.Server.ApiPrefix

	public := router.Group(apiPrefix)
	{
		public.GET("/captcha", authController.GetCaptcha)
		public.POST("/login", authController.Login)
	}

	protected := router.Group(apiPrefix)
	protected.Use(middleware.AuthMiddleware())
	{
		protected.POST("/logout", authController.Logout)
		protected.GET("/profile", authController.GetProfile)
	}
}
