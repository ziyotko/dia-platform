package controllers

import (
	"member/internal/service"
	"member/pkg/response"

	"github.com/gin-gonic/gin"
)

type SystemConfigController struct {
	systemConfigService service.SystemConfigService
}

// List returns all system configs
func (ctrl *SystemConfigController) List(c *gin.Context) {
	list, err := ctrl.systemConfigService.List()
	if err != nil {
		response.ServerErrorFrom(c, err)
		return
	}
	response.Success(c, list)
}

// Create creates a new system config
func (ctrl *SystemConfigController) Create(c *gin.Context) {
	var req service.SystemConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	cfg, err := ctrl.systemConfigService.Create(req)
	if err != nil {
		response.BadRequestFrom(c, err)
		return
	}
	response.SuccessWithMessage(c, "创建成功", cfg)
}

// Update updates a system config
func (ctrl *SystemConfigController) Update(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req service.UpdateSystemConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.systemConfigService.Update(id, req); err != nil {
		response.BadRequestFrom(c, err)
		return
	}
	response.SuccessWithMessage(c, "保存成功", nil)
}

// Delete removes a system config
func (ctrl *SystemConfigController) Delete(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.systemConfigService.Delete(id); err != nil {
		response.BadRequestFrom(c, err)
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}
