package controllers

import (
	"conference/internal/middleware"
	"conference/internal/service"
	"conference/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService service.AuthService
}

// --- Captcha ---

func (ctrl *AuthController) GetCaptcha(c *gin.Context) {
	id, b64s, _, err := ctrl.authService.GenerateCaptcha()
	if err != nil {
		response.Fail(c, "生成验证码失败")
		return
	}
	response.Ok(c, gin.H{
		"captchaId":    id,
		"captchaImage": b64s,
	})
}

// --- Member Auth ---

type MemberLoginReq struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	CaptchaID   string `json:"captchaId" binding:"required"`
	CaptchaCode string `json:"captchaCode" binding:"required"`
}

func (ctrl *AuthController) MemberLogin(c *gin.Context) {
	var req MemberLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整信息")
		return
	}
	data, err := ctrl.authService.MemberLogin(req.Username, req.Password, req.CaptchaID, req.CaptchaCode)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, data)
}

type MemberRegisterReq struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	RealName    string `json:"realName" binding:"required"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Company     string `json:"company"`
	Branch      string `json:"branch"`
	MemberLevel string `json:"memberLevel"`
}

func (ctrl *AuthController) MemberRegister(c *gin.Context) {
	var req service.MemberRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整信息")
		return
	}
	err := ctrl.authService.MemberRegister(req)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "注册成功", nil)
}

func (ctrl *AuthController) GetMemberProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := ctrl.authService.GetMemberProfile(userID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, user)
}

func (ctrl *AuthController) UpdateMemberProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	err := ctrl.authService.UpdateMemberProfile(userID, req)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "更新成功", nil)
}

func (ctrl *AuthController) ChangeMemberPassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整信息")
		return
	}
	err := ctrl.authService.ChangeMemberPassword(userID, req.OldPassword, req.NewPassword)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "密码修改成功", nil)
}

// --- Admin Auth ---

type AdminLoginReq struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	CaptchaID   string `json:"captchaId" binding:"required"`
	CaptchaCode string `json:"captchaCode" binding:"required"`
}

func (ctrl *AuthController) AdminLogin(c *gin.Context) {
	var req AdminLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整信息")
		return
	}
	data, err := ctrl.authService.AdminLogin(req.Username, req.Password, req.CaptchaID, req.CaptchaCode)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, data)
}

func (ctrl *AuthController) GetAdminProfile(c *gin.Context) {
	adminID := middleware.GetAdminID(c)
	admin, err := ctrl.authService.GetAdminProfile(adminID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, admin)
}

func (ctrl *AuthController) UpdateAdminProfile(c *gin.Context) {
	adminID := middleware.GetAdminID(c)
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	err := ctrl.authService.UpdateAdminProfile(adminID, req)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "更新成功", nil)
}

func (ctrl *AuthController) ChangeAdminPassword(c *gin.Context) {
	adminID := middleware.GetAdminID(c)
	var req struct {
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整信息")
		return
	}
	err := ctrl.authService.ChangeAdminPassword(adminID, req.OldPassword, req.NewPassword)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "密码修改成功", nil)
}
