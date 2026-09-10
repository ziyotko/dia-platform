package controllers

import (
	"application/internal/models"
	"application/internal/service"
	"application/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	service service.UserService
}

// --- Applicant management (申报人) ---

func (ctrl *UserController) ListUsers(c *gin.Context) {
	page, size := getPage(c)
	keyword := c.Query("keyword")
	list, total, err := ctrl.service.ListUsers(page, size, keyword)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctrl *UserController) CreateUser(c *gin.Context) {
	var u models.User
	if err := c.ShouldBindJSON(&u); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.CreateUser(&u); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "用户管理", "新增申报人", u.Username)
	response.Ok(c, u)
}

func (ctrl *UserController) UpdateUser(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.UpdateUser(id, updates); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "用户管理", "编辑申报人", "id="+c.Param("id"))
	response.OkWithMessage(c, "更新成功", nil)
}

func (ctrl *UserController) DeleteUser(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.service.DeleteUser(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "用户管理", "删除申报人", "id="+c.Param("id"))
	response.OkWithMessage(c, "删除成功", nil)
}

type SetStatusReq struct {
	Status int `json:"status"`
}

func (ctrl *UserController) SetUserStatus(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req SetStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.SetUserStatus(id, req.Status); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "状态已更新", nil)
}

// --- Admin management (管理人 / 评审人) ---

func (ctrl *UserController) ListAdmins(c *gin.Context) {
	page, size := getPage(c)
	keyword := c.Query("keyword")
	list, total, err := ctrl.service.ListAdmins(page, size, keyword)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctrl *UserController) CreateAdmin(c *gin.Context) {
	var a models.Admin
	if err := c.ShouldBindJSON(&a); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.CreateAdmin(&a); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "账号管理", "新增管理账号", a.Username)
	response.Ok(c, a)
}

func (ctrl *UserController) UpdateAdmin(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.UpdateAdmin(id, updates); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "账号管理", "编辑管理账号", "id="+c.Param("id"))
	response.OkWithMessage(c, "更新成功", nil)
}

func (ctrl *UserController) DeleteAdmin(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.service.DeleteAdmin(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "账号管理", "删除管理账号", "id="+c.Param("id"))
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctrl *UserController) ListRoles(c *gin.Context) {
	list, err := ctrl.service.ListRoles()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, list)
}
