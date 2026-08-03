package controllers

import (
	"conference/internal/middleware"
	"conference/internal/service"
	"conference/pkg/response"

	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	service service.DashboardService
}

func (ctrl *DashboardController) GetAdminStats(c *gin.Context) {
	stats, err := ctrl.service.GetAdminStats()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, stats)
}

func (ctrl *DashboardController) GetMemberStats(c *gin.Context) {
	userID := middleware.GetUserID(c)
	stats, err := ctrl.service.GetMemberStats(userID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, stats)
}
