package controllers

import (
	"base/internal/service"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	service service.DashboardService
}

// Stats 仪表盘统计。传入当前用户与是否租户管理员，用于「我的」类指标与普通用户的范围收敛。
func (ctl *DashboardController) Stats(c *gin.Context) {
	userID := c.GetUint64("userID")
	tenantID := c.GetUint64("tenantID")
	stats, err := ctl.service.GetStats(tenantID, userID, isAdminUser(userID))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, stats)
}
