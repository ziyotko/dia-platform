package controllers

import (
	"strconv"

	"conference/internal/models"
	"conference/internal/service"
	"conference/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	service service.UserService
}

// --- Admin: Member user management ---

func (ctrl *UserController) Create(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.Create(&user); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, user)
}

func (ctrl *UserController) Update(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.Update(id, updates); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "更新成功", nil)
}

func (ctrl *UserController) Delete(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.service.Delete(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctrl *UserController) GetByID(c *gin.Context) {
	id := parseUint(c.Param("id"))
	user, err := ctrl.service.GetByID(id)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, user)
}

func (ctrl *UserController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")
	branch := c.Query("branch")
	level := c.Query("level")

	list, total, err := ctrl.service.List(keyword, branch, level, page, size)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctrl *UserController) SetValidity(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req struct {
		IsValid bool `json:"isValid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.SetValidity(id, req.IsValid); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "设置成功", nil)
}

// --- Admin: Admin management ---

func (ctrl *UserController) CreateAdmin(c *gin.Context) {
	var admin models.Admin
	if err := c.ShouldBindJSON(&admin); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.CreateAdmin(&admin); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, admin)
}

func (ctrl *UserController) ListAdmins(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	list, total, err := ctrl.service.ListAdmins(page, size)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
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
	response.OkWithMessage(c, "更新成功", nil)
}

// --- Admin: Roles ---

func (ctrl *UserController) ListRoles(c *gin.Context) {
	roles, err := ctrl.service.ListRoles()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, roles)
}
