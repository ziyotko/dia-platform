package controllers

import (
	"application/internal/middleware"
	"application/internal/service"
	"application/pkg/response"
	"application/pkg/utils"

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
	// 字段名与 portal 保持一致：captcha_id + captcha_img
	response.Ok(c, gin.H{
		"captcha_id":  id,
		"captcha_img": b64s,
	})
}

// --- Applicant Auth (申报人) ---

func (ctrl *AuthController) UserLogin(c *gin.Context) {
	var req struct {
		Username    string `json:"username" binding:"required"`
		Password    string `json:"password" binding:"required"`
		CaptchaID   string `json:"captcha_id" binding:"required"`
		CaptchaCode string `json:"captcha_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整信息")
		return
	}
	data, err := ctrl.authService.UserLogin(req.Username, req.Password, req.CaptchaID, req.CaptchaCode)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, data)
}

func (ctrl *AuthController) UserRegister(c *gin.Context) {
	var req service.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整信息（含验证码）")
		return
	}
	if err := ctrl.authService.UserRegister(req); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "注册成功", nil)
}

func (ctrl *AuthController) GetUserProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := ctrl.authService.GetUserProfile(userID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, user)
}

func (ctrl *AuthController) UpdateUserProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.authService.UpdateUserProfile(userID, req); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "更新成功", nil)
}

func (ctrl *AuthController) ChangeUserPassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整信息")
		return
	}
	if err := ctrl.authService.ChangeUserPassword(userID, req.OldPassword, req.NewPassword); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "密码修改成功", nil)
}

// UserLogout 申报人退出登录：把当前 Token 的 jti 写入黑名单（TTL = Token 剩余有效期），
// 使其在自然过期前不可再用。写黑名单失败不影响本地登出，只记日志。
func (ctrl *AuthController) UserLogout(c *gin.Context) {
	if err := ctrl.authService.Logout(middleware.GetTokenID(c), middleware.GetTokenExpiry(c)); err != nil {
		utils.LogWarn("申报人退出登录写黑名单失败: %s", err)
	}
	response.OkWithMessage(c, "已退出登录", nil)
}

// --- Admin Auth (管理人 / 评审人) ---

func (ctrl *AuthController) AdminLogin(c *gin.Context) {
	var req struct {
		Username    string `json:"username" binding:"required"`
		Password    string `json:"password" binding:"required"`
		CaptchaID   string `json:"captcha_id" binding:"required"`
		CaptchaCode string `json:"captcha_code" binding:"required"`
	}
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
	if err := ctrl.authService.UpdateAdminProfile(adminID, req); err != nil {
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
	if err := ctrl.authService.ChangeAdminPassword(adminID, req.OldPassword, req.NewPassword); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "密码修改成功", nil)
}

// AdminLogout 管理端退出登录：把当前 Token 的 jti 写入黑名单（TTL = Token 剩余有效期）。
func (ctrl *AuthController) AdminLogout(c *gin.Context) {
	if err := ctrl.authService.Logout(middleware.GetTokenID(c), middleware.GetTokenExpiry(c)); err != nil {
		utils.LogWarn("管理员退出登录写黑名单失败: %s", err)
	}
	response.OkWithMessage(c, "已退出登录", nil)
}
