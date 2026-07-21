package controllers

import (
	"base/internal/models"
	"base/internal/service"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

type OrganizationController struct {
	service service.OrganizationService
}

func (ctl *OrganizationController) Create(c *gin.Context) {
	var o models.Organization
	if err := c.ShouldBindJSON(&o); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	o.TenantID = c.GetUint64("tenantID")
	if err := ctl.service.Create(&o); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, o)
}

func (ctl *OrganizationController) Update(c *gin.Context) {
	var o models.Organization
	if err := c.ShouldBindJSON(&o); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	o.ID = uint64(parseID(c))
	if err := ctl.service.Update(&o); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, o)
}

func (ctl *OrganizationController) Delete(c *gin.Context) {
	id := uint64(parseID(c))
	if err := ctl.service.Delete(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctl *OrganizationController) Get(c *gin.Context) {
	id := uint64(parseID(c))
	o, err := ctl.service.GetByID(id)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, o)
}

func (ctl *OrganizationController) Tree(c *gin.Context) {
	tenantID := c.GetUint64("tenantID")
	orgs, err := ctl.service.GetTree(tenantID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, orgs)
}
