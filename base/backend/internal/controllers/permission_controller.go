package controllers

import (
	"base/internal/models"
	"base/internal/service"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

type PermissionController struct {
	service service.PermissionService
}

func (ctl *PermissionController) Create(c *gin.Context) {
	var p models.Permission
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	if err := ctl.service.Create(&p); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, p)
}

func (ctl *PermissionController) Update(c *gin.Context) {
	var p models.Permission
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	p.ID = uint64(parseID(c))
	if err := ctl.service.Update(&p); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, p)
}

func (ctl *PermissionController) Delete(c *gin.Context) {
	id := uint64(parseID(c))
	if err := ctl.service.Delete(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctl *PermissionController) Tree(c *gin.Context) {
	appCode := c.DefaultQuery("appCode", "")
	menus, err := ctl.service.GetTree(appCode)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, menus)
}
