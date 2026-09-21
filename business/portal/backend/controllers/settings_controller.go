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

// maskedSecret 敏感字段的脱敏占位值：读取设置时返回该值，提交时若原样回传则不更新该字段。
const maskedSecret = "******"

func NewSettingsController() *SettingsController {
	return &SettingsController{
		settingsService: &services.SettingsService{},
	}
}

func (c *SettingsController) GetMinPasswordLengthSettings(ctx *gin.Context) {
	settings, err := c.settingsService.GetMinPasswordLengthSettings()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取设置失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取设置成功", settings))
}

func (c *SettingsController) GetSettings(ctx *gin.Context) {
	settings, err := c.settingsService.GetSettings()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取设置失败"))
		return
	}
	// 邮件授权码等敏感字段不返回明文，避免抓包/前端缓存泄露（未配置时返回空）
	if settings.EmailPassword != "" {
		settings.EmailPassword = maskedSecret
	}
	ctx.JSON(http.StatusOK, utils.Success("获取设置成功", settings))
}

// TestEmail 邮件（SMTP）连接测试：按当前设置尝试建立连接与认证，不发送邮件
func (c *SettingsController) TestEmail(ctx *gin.Context) {
	if err := c.settingsService.TestEmailConnection(); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "邮件连接测试失败: "+utils.SafeErrText(err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("邮件连接测试成功", nil))
}

func (c *SettingsController) GetPublicSiteInfo(ctx *gin.Context) {
	settings, err := c.settingsService.GetSettings()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Success("获取成功", gin.H{
			"siteName":       "门户网站管理后台",
			"siteUrl":        "",
			"logo":           "",
			"icp":            "",
			"copyright":      "门户网站管理系统 版权所有",
			"captchaEnabled": true,
		}))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取成功", gin.H{
		"siteName":       settings.SiteName,
		"siteUrl":        settings.SiteUrl,
		"logo":           settings.Logo,
		"icp":            settings.Icp,
		"copyright":      settings.Copyright,
		"captchaEnabled": settings.CaptchaEnabled,
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
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}

	var req models.Setting
	if err := json.Unmarshal(body, &req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}

	// 敏感字段原样回传脱敏占位值时，视为未修改，不参与更新
	if req.EmailPassword == maskedSecret {
		delete(raw, "emailPassword")
	}

	fields := settingFieldsFromKeys(raw)
	if len(fields) == 0 {
		ctx.JSON(http.StatusOK, utils.Error(1, "没有可更新的设置项"))
		return
	}

	if err := c.settingsService.UpdateSettings(&req, fields...); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("保存设置失败", err)))
		return
	}

	// 设置更新后同步刷新静态化参数缓存，保证读取接口返回最新值；
	// 刷新失败时清除缓存，下次读取自动回源，避免 DB/缓存长期不一致
	if err := c.settingsService.LoadStaticParamsToCache(); err != nil {
		utils.Logger.Warnf("保存设置后刷新静态化参数缓存失败，已清除缓存待下次回源: %s", err)
		if delErr := c.settingsService.InvalidateStaticParamsCache(); delErr != nil {
			utils.Logger.Warnf("清除静态化参数缓存失败: %s", delErr)
		}
	}

	ctx.JSON(http.StatusOK, utils.Success("保存设置成功", nil))
}

// settingFieldsFromKeys 根据请求体中出现的 JSON key，返回对应 Setting 结构体的字段名列表
func settingFieldsFromKeys(raw map[string]json.RawMessage) []string {
	jsonToField := make(map[string]string)
	t := reflect.TypeOf(models.Setting{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous { // 兜底：跳过内嵌结构体字段（Setting 的字段均平铺定义）
			continue
		}
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		name := strings.Split(tag, ",")[0]
		jsonToField[name] = field.Name
	}

	// 主键与时间戳等系统字段禁止通过请求体覆盖
	// （Setting 的字段均平铺定义，Anonymous 恒为 false，需显式排除）
	skipFields := map[string]bool{
		"id": true, "createdAt": true, "updatedAt": true,
	}

	fields := make([]string, 0, len(raw))
	for key := range raw {
		if skipFields[key] {
			continue
		}
		if fieldName, ok := jsonToField[key]; ok {
			fields = append(fields, fieldName)
		}
	}
	return fields
}
