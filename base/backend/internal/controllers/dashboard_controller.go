package controllers

import (
	"base/internal/service"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	service service.DashboardService
}

func (ctl *DashboardController) Stats(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	stats, err := ctl.service.GetStats(tenantID.(uint64))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, stats)
}
