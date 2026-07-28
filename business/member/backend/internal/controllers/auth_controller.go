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

// DownloadApplicationTemplate serves the membership application form template
func (ctrl *AuthController) DownloadApplicationTemplate(c *gin.Context) {
	content := `中国电器工业协会入会申请表

申请日期：______年____月____日

一、申请单位信息
单位名称：__________________________
统一社会信用代码：__________________
法定代表人：______________ 职务：____
联系人：__________________ 电话：____
手机：__________________ 邮箱：______
单位地址：__________________________
邮政编码：______________
单位网址：__________________________

二、申请会员类别
□ 单位会员  □ 个人会员

三、申请分会
□ 变压器分会  □ 高压开关分会  □ 电线电缆分会
□ 电机分会    □ 新能源电器分会  □ 绝缘材料分会
□ 电力电子分会  □ 其他：__________

四、单位简介
_______________________________________________
_______________________________________________
_______________________________________________

五、申请单位意见（签字盖章）
法定代表人签字：______________
单位盖章：
日期：______年____月____日

六、协会审批意见
审批人：______________
协会盖章：
日期：______年____月____日

注：请将此表打印并签字盖章后，上传至会员系统。`
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=入会申请表.txt")
	c.String(200, content)
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
