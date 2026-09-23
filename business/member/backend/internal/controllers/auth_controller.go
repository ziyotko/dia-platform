package controllers

import (
	"path/filepath"
	"strings"

	"member/config"
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
		"captcha_id":  id,
		"captcha_img": b64s,
	})
}

// CheckExists checks whether a username/mobile/email is already registered
func (ctrl *AuthController) CheckExists(c *gin.Context) {
	var req service.CheckExistsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	exists, err := ctrl.authService.CheckExists(req.Field, req.Value)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"exists": exists})
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

// 允许上传的文件扩展名白名单
var allowedUploadExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
	".webp": true, ".bmp": true, ".pdf": true,
	".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".txt": true,
	// 入会申请页明确提示「多页扫描件可打包为 ZIP/RAR」，这里必须同步放行，
	// 否则按页面指引上传压缩包会直接 400（前端 accept 也只列了这些）
	".zip": true, ".rar": true,
}

// 允许的上传子目录白名单，防止路径穿越
var allowedUploadDirs = map[string]bool{
	"files": true, "avatars": true, "certificates": true, "cert": true,
	"charter": true, "invoices": true, "images": true, "articles": true,
	"certs": true, "covers": true, "templates": true,
}

const maxUploadSize = 10 << 20 // 10MB

// UploadFile handles file upload
func (ctrl *AuthController) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择文件")
		return
	}

	// 限制文件大小，防止磁盘耗尽
	if file.Size > maxUploadSize {
		response.BadRequest(c, "文件大小不能超过 10MB")
		return
	}

	// 扩展名白名单
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedUploadExts[ext] {
		response.BadRequest(c, "不支持的文件类型")
		return
	}

	// 子目录白名单，阻断路径穿越
	subDir := c.DefaultPostForm("dir", "files")
	if !allowedUploadDirs[subDir] {
		response.BadRequest(c, "非法上传目录")
		return
	}

	path, err := utils.SaveUploadedFile(file, subDir)
	if err != nil {
		response.ServerError(c, "文件上传失败")
		return
	}

	response.Success(c, gin.H{
		"url":  config.Cfg.Server.UploadDirPrefix + "/" + path,
		"name": file.Filename,
	})
}
