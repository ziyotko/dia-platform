package controllers

import (
	"base/internal/models"
	"base/internal/service"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService service.AuthService
}

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResp struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

func (ctl *AuthController) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	user, token, err := ctl.authService.Login(req.Username, req.Password)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, LoginResp{Token: token, User: user})
}

func (ctl *AuthController) Info(c *gin.Context) {
	userID := c.GetUint64("userID")
	user, err := ctl.authService.GetUserInfo(userID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, user)
}

func (ctl *AuthController) Menus(c *gin.Context) {
	userID := c.GetUint64("userID")
	tenantID := c.GetUint64("tenantID")
	menus, err := service.MenuService{}.GetUserMenus(userID, tenantID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, menus)
}

func (ctl *AuthController) Permissions(c *gin.Context) {
	userID := c.GetUint64("userID")
	perms, err := ctl.authService.GetUserPermissions(userID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, perms)
}

func (ctl *AuthController) InitAdmin(c *gin.Context) {
	var req struct {
		Password string `json:"password"`
	}
	c.ShouldBindJSON(&req)
	if req.Password == "" {
		req.Password = "admin123"
	}
	if err := ctl.authService.InitSuperAdmin(req.Password); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "初始化成功，账号 admin", nil)
}

func (ctl *AuthController) ChangePassword(c *gin.Context) {
	var req struct {
		OldPwd string `json:"oldPwd" binding:"required"`
		NewPwd string `json:"newPwd" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	userID := c.GetUint64("userID")
	userService := service.UserService{}
	if err := userService.ChangePassword(userID, req.OldPwd, req.NewPwd); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "密码修改成功", nil)
}
