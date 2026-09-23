package controllers

import (
	"member/internal/middleware"
	"member/internal/service"
	"member/pkg/response"

	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	dashService service.DashboardService
}

func (ctrl *DashboardController) GetMemberDashboard(c *gin.Context) {
	memberID := middleware.GetMemberID(c)
	dash, err := ctrl.dashService.GetMemberDashboard(memberID)
	if err != nil {
		response.ServerErrorFrom(c, err)
		return
	}
	response.Success(c, dash)
}
