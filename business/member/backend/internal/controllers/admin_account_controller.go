package controllers

import (
	"member/internal/middleware"
	"member/internal/service"
	"member/pkg/response"

	"github.com/gin-gonic/gin"
)

type AdminAccountController struct {
	adminAccountService service.AdminAccountService
}

func (ctrl *AdminAccountController) List(c *gin.Context) {
	accounts, err := ctrl.adminAccountService.List()
	if err != nil {
		response.ServerErrorFrom(c, err)
		return
	}
	response.Success(c, accounts)
}

func (ctrl *AdminAccountController) Create(c *gin.Context) {
	var req service.AdminAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写用户名")
		return
	}
	result, err := ctrl.adminAccountService.Create(req)
	if err != nil {
		response.BadRequestFrom(c, err)
		return
	}
	response.SuccessWithMessage(c, "下属管理员已创建", result)
}

func (ctrl *AdminAccountController) Update(c *gin.Context) {
	var req service.AdminAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.adminAccountService.Update(parseUint(c.Param("id")), req); err != nil {
		response.BadRequestFrom(c, err)
		return
	}
	response.SuccessWithMessage(c, "下属管理员信息已更新", nil)
}

func (ctrl *AdminAccountController) ResetPassword(c *gin.Context) {
	password, err := ctrl.adminAccountService.ResetPassword(parseUint(c.Param("id")))
	if err != nil {
		response.BadRequestFrom(c, err)
		return
	}
	response.SuccessWithMessage(c, "密码已重置", gin.H{"password": password})
}

func (ctrl *AdminAccountController) Delete(c *gin.Context) {
	if err := ctrl.adminAccountService.Delete(parseUint(c.Param("id")), middleware.GetMemberID(c)); err != nil {
		response.BadRequestFrom(c, err)
		return
	}
	response.SuccessWithMessage(c, "下属管理员已删除", nil)
}
