package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/services"
	"server/utils"
)

type SettingsController struct {
	settingsService *services.SettingsService
}

func NewSettingsController() *SettingsController {
	return &SettingsController{
		settingsService: &services.SettingsService{},
	}
}

func (c *SettingsController) GetSettings(ctx *gin.Context) {
	settings, err := c.settingsService.GetSettings()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取设置失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取设置成功", settings))
}

func (c *SettingsController) GetPublicSiteInfo(ctx *gin.Context) {
	settings, err := c.settingsService.GetSettings()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Success("获取成功", gin.H{
			"siteName":  "门户网站管理后台",
			"logo":      "",
			"icp":       "",
			"copyright": "门户网站管理系统 版权所有",
		}))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取成功", gin.H{
		"siteName":  settings.SiteName,
		"logo":      settings.Logo,
		"icp":       settings.Icp,
		"copyright": settings.Copyright,
	}))
}

func (c *SettingsController) UpdateSettings(ctx *gin.Context) {
	var req models.Settings
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: "+err.Error()))
		return
	}

	err := c.settingsService.UpdateSettings(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "保存设置失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("保存设置成功", nil))
}
