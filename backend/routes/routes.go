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

		protected.GET("/users", userController.GetUsers)
		protected.POST("/users", userController.CreateUser)
		protected.GET("/users/:id", userController.GetUserByID)
		protected.PUT("/users/:id", userController.UpdateUser)
		protected.DELETE("/users/:id", userController.DeleteUser)
		protected.PATCH("/users/:id/status", userController.UpdateUserStatus)
	}
}
