package controllers

import (
	"strconv"

	"base/internal/models"
	"base/internal/service"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

type RoleController struct {
	service service.RoleService
}

func (ctl *RoleController) Create(c *gin.Context) {
	var r models.Role
	if err := c.ShouldBindJSON(&r); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	r.TenantID = c.GetUint64("tenantID")
	if err := ctl.service.Create(&r); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, r)
}

func (ctl *RoleController) Update(c *gin.Context) {
	var r models.Role
	if err := c.ShouldBindJSON(&r); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	r.ID = uint64(parseID(c))
	if err := ctl.service.Update(&r); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, r)
}

func (ctl *RoleController) Delete(c *gin.Context) {
	id := uint64(parseID(c))
	if err := ctl.service.Delete(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctl *RoleController) Get(c *gin.Context) {
	id := uint64(parseID(c))
	r, err := ctl.service.GetByID(id)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, r)
}

func (ctl *RoleController) List(c *gin.Context) {
	tenantID := c.GetUint64("tenantID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	keyword := c.Query("keyword")
	list, total, err := ctl.service.List(tenantID, page, size, keyword)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctl *RoleController) AssignMenus(c *gin.Context) {
	id := uint64(parseID(c))
	var req struct {
		MenuIDs []uint64 `json:"menuIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	if err := ctl.service.AssignMenus(id, req.MenuIDs); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "分配成功", nil)
}

func (ctl *RoleController) AssignPermissions(c *gin.Context) {
	id := uint64(parseID(c))
	var req struct {
		PermissionIDs []uint64 `json:"permissionIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	if err := ctl.service.AssignPermissions(id, req.PermissionIDs); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "分配成功", nil)
}
