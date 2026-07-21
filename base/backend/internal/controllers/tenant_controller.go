package controllers

import (
	"strconv"

	"base/internal/models"
	"base/internal/service"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

type TenantController struct {
	service service.TenantService
}

func (ctl *TenantController) Create(c *gin.Context) {
	var t models.Tenant
	if err := c.ShouldBindJSON(&t); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	if err := ctl.service.Create(&t); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, t)
}

func (ctl *TenantController) Update(c *gin.Context) {
	var t models.Tenant
	if err := c.ShouldBindJSON(&t); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	t.ID = uint64(parseID(c))
	if err := ctl.service.Update(&t); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, t)
}

func (ctl *TenantController) Delete(c *gin.Context) {
	id := parseID(c)
	if err := ctl.service.Delete(uint64(id)); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctl *TenantController) Get(c *gin.Context) {
	id := uint64(parseID(c))
	t, err := ctl.service.GetByID(id)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, t)
}

func (ctl *TenantController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	keyword := c.Query("keyword")
	list, total, err := ctl.service.List(page, size, keyword)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func parseID(c *gin.Context) int {
	id, _ := strconv.Atoi(c.Param("id"))
	return id
}
