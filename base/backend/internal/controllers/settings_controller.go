package controllers

import (
	"base/internal/models"
	"base/internal/service"
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
