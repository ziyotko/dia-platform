package controllers

import (
	"strconv"

	"base/internal/models"
	"base/internal/service"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	service service.UserService
}

func (ctl *UserController) Create(c *gin.Context) {
	var u models.User
	if err := c.ShouldBindJSON(&u); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	u.TenantID = c.GetUint64("tenantID")
	if err := ctl.service.Create(&u); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, u)
}

func (ctl *UserController) Update(c *gin.Context) {
	var u models.User
	if err := c.ShouldBindJSON(&u); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	u.ID = uint64(parseID(c))
	if err := ctl.service.Update(&u); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, u)
}

func (ctl *UserController) Delete(c *gin.Context) {
	id := uint64(parseID(c))
	if err := ctl.service.Delete(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctl *UserController) Get(c *gin.Context) {
	id := uint64(parseID(c))
	u, err := ctl.service.GetByID(id)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, u)
}

func (ctl *UserController) List(c *gin.Context) {
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

func (ctl *UserController) AssignRoles(c *gin.Context) {
	id := uint64(parseID(c))
	var req struct {
		RoleIDs []uint64 `json:"roleIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	if err := ctl.service.AssignRoles(id, req.RoleIDs); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "分配成功", nil)
}

func (ctl *UserController) ResetPassword(c *gin.Context) {
	id := uint64(parseID(c))
	if err := ctl.service.ResetPassword(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "密码已重置为 123456", nil)
}
