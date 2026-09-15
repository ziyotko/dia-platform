package controllers

import (
	"base/internal/models"
	"base/internal/service"
	"base/pkg/notifier"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

type SettingsController struct {
	service service.SettingsService
}

func (ctl *SettingsController) Get(c *gin.Context) {
	category := c.Query("category")
	var data interface{}
	var err error
	if category != "" {
		data, err = ctl.service.GetByCategory(category)
	} else {
		data, err = ctl.service.GetAll()
	}
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, data)
}

func (ctl *SettingsController) Save(c *gin.Context) {
	var req struct {
		Settings []models.Setting `json:"settings" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	if err := ctl.service.BatchSave(req.Settings); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "保存成功", nil)
}

// SiteInfo 公开站点信息（无需登录）：登录页据此决定是否展示验证码等。
// 只返回与登录页展示相关的非敏感设置。
func (ctl *SettingsController) SiteInfo(c *gin.Context) {
	security := ctl.service.GetSecuritySettings()
	response.Ok(c, gin.H{
		"captchaEnabled": security.CaptchaEnabled,
	})
}

func (ctl *SettingsController) TestEmail(c *gin.Context) {
	var req struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		From     string `json:"from"`
		SSL      bool   `json:"ssl"`
		To       string `json:"to" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}

	sender := notifier.NewEmailSender(notifier.EmailConfig{
		Host:     req.Host,
		Port:     req.Port,
		Username: req.Username,
		Password: req.Password,
		From:     req.From,
		SSL:      req.SSL,
	})
	if err := sender.Send(notifier.Message{
		To:      req.To,
		Subject: "Base 平台邮件测试",
		Body:    "这是一封来自 Base 底座平台的测试邮件，如果您的邮箱能收到，说明邮件配置正确。",
	}); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "测试邮件已发送", nil)
}
