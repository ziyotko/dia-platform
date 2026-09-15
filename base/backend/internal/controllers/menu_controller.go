package controllers

import (
	"base/internal/models"
	"base/internal/service"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

type MenuController struct {
	service service.MenuService
}

func (ctl *MenuController) Create(c *gin.Context) {
	var m models.Menu
	if err := c.ShouldBindJSON(&m); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	// 与其它模块统一：普通租户用户强制落到自身租户；平台超管可用 body 的 tenantId 指定（缺省=平台内置菜单）
	m.TenantID = resolveTenantID(c, m.TenantID)
	if err := ctl.service.Create(&m); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, m)
}

func (ctl *MenuController) Update(c *gin.Context) {
	var m models.Menu
	if err := c.ShouldBindJSON(&m); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	m.ID = uint64(parseID(c))
	tenantID := c.GetUint64("tenantID")
	if err := ctl.service.Update(&m, tenantID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, m)
}

func (ctl *MenuController) Delete(c *gin.Context) {
	id := uint64(parseID(c))
	tenantID := c.GetUint64("tenantID")
	if err := ctl.service.Delete(id, tenantID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctl *MenuController) Tree(c *gin.Context) {
	tenantID := c.GetUint64("tenantID")
	appCode := c.DefaultQuery("appCode", "")
	menus, err := ctl.service.GetTree(tenantID, appCode)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, menus)
}
