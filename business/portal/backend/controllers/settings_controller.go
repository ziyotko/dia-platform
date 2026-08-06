package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"

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
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误"))
		return
	}

	// 解析请求体中实际包含的字段，只更新这些字段，避免覆盖表中其他设置项
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: "+err.Error()))
		return
	}

	var req models.Setting
	if err := json.Unmarshal(body, &req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: "+err.Error()))
		return
	}

	fields := settingFieldsFromKeys(raw)
	if len(fields) == 0 {
		ctx.JSON(http.StatusOK, utils.Error(1, "没有可更新的设置项"))
		return
	}

	if err := c.settingsService.UpdateSettings(&req, fields...); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "保存设置失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("保存设置成功", nil))
}

// settingFieldsFromKeys 根据请求体中出现的 JSON key，返回对应 Setting 结构体的字段名列表
func settingFieldsFromKeys(raw map[string]json.RawMessage) []string {
	jsonToField := make(map[string]string)
	t := reflect.TypeOf(models.Setting{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous { // 跳过 gorm.Model 内嵌字段（id/created_at/updated_at/deleted_at）
			continue
		}
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		name := strings.Split(tag, ",")[0]
		jsonToField[name] = field.Name
	}

	fields := make([]string, 0, len(raw))
	for key := range raw {
		if fieldName, ok := jsonToField[key]; ok {
			fields = append(fields, fieldName)
		}
	}
	return fields
}
