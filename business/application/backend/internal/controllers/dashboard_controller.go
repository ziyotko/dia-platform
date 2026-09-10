package controllers

import (
	"application/internal/middleware"
	"application/internal/models"
	"application/internal/service"
	"application/pkg/response"

	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	service service.DashboardService
}

func (ctrl *DashboardController) UserStats(c *gin.Context) {
	userID := middleware.GetUserID(c)
	stats := ctrl.service.UserStats(userID)
	response.Ok(c, stats)
}

func (ctrl *DashboardController) AdminStats(c *gin.Context) {
	roleCode := middleware.GetRoleCode(c)
	adminID := middleware.GetAdminID(c)
	if roleCode == models.RoleReviewer {
		stats := ctrl.service.ReviewerStats(adminID)
		response.Ok(c, stats)
		return
	}
	stats := ctrl.service.AdminStats()
	response.Ok(c, stats)
}
