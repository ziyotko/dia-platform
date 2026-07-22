package controllers

import (
	"strconv"

	"base/internal/models"
	"base/internal/service"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

type AppInstanceController struct {
	service service.AppInstanceService
}

func (ctl *AppInstanceController) Create(c *gin.Context) {
	var i models.AppInstance
	if err := c.ShouldBindJSON(&i); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	i.TenantID = c.GetUint64("tenantID")
	if err := ctl.service.Create(&i); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, i)
}

func (ctl *AppInstanceController) Update(c *gin.Context) {
	var i models.AppInstance
	if err := c.ShouldBindJSON(&i); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	i.ID = uint64(parseID(c))
	tenantID := c.GetUint64("tenantID")
	if err := ctl.service.Update(&i, tenantID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, i)
}

func (ctl *AppInstanceController) Delete(c *gin.Context) {
	id := uint64(parseID(c))
	tenantID := c.GetUint64("tenantID")
	if err := ctl.service.Delete(id, tenantID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctl *AppInstanceController) List(c *gin.Context) {
	tenantID := c.GetUint64("tenantID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	list, total, err := ctl.service.ListByTenant(tenantID, page, size)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctl *AppInstanceController) MyApps(c *gin.Context) {
	tenantID := c.GetUint64("tenantID")
	apps, err := ctl.service.GetTenantActiveApps(tenantID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, apps)
}
