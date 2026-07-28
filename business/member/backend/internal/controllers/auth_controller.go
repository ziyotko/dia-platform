package controllers

import (
	"member/internal/middleware"
	"member/internal/service"
	"member/pkg/captcha"
	"member/pkg/response"
	"member/pkg/utils"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService service.AuthService
}

// GetCaptcha generates a captcha
func (ctrl *AuthController) GetCaptcha(c *gin.Context) {
	id, b64s, err := captcha.Generate()
	if err != nil {
		response.ServerError(c, "验证码生成失败")
		return
	}
	response.Success(c, gin.H{
		"captcha_id":    id,
		"captcha_image": b64s,
	})
}

// Register handles member registration
func (ctrl *AuthController) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整信息")
		return
	}
	if req.MemberType == "" {
		req.MemberType = "unit"
	}

	result, err := ctrl.authService.Register(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "注册成功", result)
}

// Login handles member login
func (ctrl *AuthController) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整信息")
		return
	}

	result, err := ctrl.authService.Login(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, result)
}

// GetProfile returns the member's profile
func (ctrl *AuthController) GetProfile(c *gin.Context) {
	memberID := middleware.GetMemberID(c)
	member, err := ctrl.authService.GetProfile(memberID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, member)
}

// UpdateProfile updates the member's profile
func (ctrl *AuthController) UpdateProfile(c *gin.Context) {
	memberID := middleware.GetMemberID(c)
	var req service.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.authService.UpdateProfile(memberID, req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "修改成功", nil)
}

// ChangePassword changes the member's password
func (ctrl *AuthController) ChangePassword(c *gin.Context) {
	memberID := middleware.GetMemberID(c)
	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.authService.ChangePassword(memberID, req.OldPassword, req.NewPassword); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "密码修改成功", nil)
}

// RequestPasswordReset sends reset email
func (ctrl *AuthController) RequestPasswordReset(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入邮箱")
		return
	}
	if err := ctrl.authService.RequestPasswordReset(req.Email); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "重置链接已发送至邮箱", nil)
}

// ResetPassword resets the password
func (ctrl *AuthController) ResetPassword(c *gin.Context) {
	var req service.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.authService.ResetPassword(req.Token, req.NewPassword); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "密码重置成功", nil)
}

// GetSiteInfo returns site configuration
func (ctrl *AuthController) GetSiteInfo(c *gin.Context) {
	config := ctrl.authService.GetSiteConfig()
	response.Success(c, config)
}

// UploadFile handles file upload
func (ctrl *AuthController) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择文件")
		return
	}

	subDir := c.DefaultPostForm("dir", "files")
	path, err := utils.SaveUploadedFile(file, subDir)
	if err != nil {
		response.ServerError(c, "文件上传失败")
		return
	}

	response.Success(c, gin.H{
		"url":  "/" + path,
		"name": file.Filename,
	})
}
