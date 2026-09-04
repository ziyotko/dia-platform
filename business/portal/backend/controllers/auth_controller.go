package controllers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/services"
	"server/utils"
)

type AuthController struct {
	userService    *services.UserService
	roleService    *services.RoleService
	logService     *services.LogService
	articleService *services.ArticleService
}

func NewAuthController() *AuthController {
	return &AuthController{
		userService:    &services.UserService{},
		roleService:    &services.RoleService{},
		logService:     &services.LogService{},
		articleService: &services.ArticleService{},
	}
}

func (c *AuthController) GetCaptcha(ctx *gin.Context) {
	id, b64s, err := utils.GenerateCaptcha()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "生成验证码失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取验证码成功", gin.H{"captcha_id": id, "captcha_img": b64s}))
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req struct {
		Email       string `json:"email"`
		Account     string `json:"account"`
		Mobile      string `json:"mobile"`
		Password    string `json:"password" binding:"required"`
		CaptchaID   string `json:"captcha_id" binding:"required"`
		CaptchaCode string `json:"captcha_code" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误"))
		return
	}

	user, token, signKey, err := c.userService.Login(req.Email, req.Account, req.Mobile, req.Password, req.CaptchaID, req.CaptchaCode)

	username := req.Account
	if username == "" {
		username = req.Email
		if username == "" {
			username = req.Mobile
		}
	}

	browser, os, device := utils.ParseUA(ctx.Request.UserAgent())
	loginLog := &models.LoginLog{
		Username: username,
		IP:       ctx.ClientIP(),
		Browser:  browser,
		OS:       os,
		Device:   device,
	}

	if err != nil {
		loginLog.Status = 0
		utils.DB.Create(loginLog)
		// 安全考虑：登录失败一律返回统一提示，避免泄露具体错误原因
		ctx.JSON(http.StatusOK, utils.Error(1, "登录失败，请稍后重试"))
		return
	}

	loginLog.Username = user.Username
	if loginLog.Username == "" {
		loginLog.Username = user.Account
	}
	loginLog.Status = 1
	utils.DB.Create(loginLog)

	// 登录返回的 user 仅暴露前端所需字段，roleIds 统一为数字数组（原 models.User.RoleIds 是逗号分隔字符串）
	userRoles, err := c.userService.GetUserRoleIds(user.ID)
	if err != nil {
		userRoles = []int{}
	}
	ctx.JSON(http.StatusOK, utils.Success("登录成功", gin.H{
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"avatar":   user.Avatar,
			"roleIds":  userRoles,
		},
		"token":   token,
		"signKey": signKey,
	}))
}

func (c *AuthController) Logout(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.JSON(http.StatusOK, utils.Error(1, "未提供token"))
		return
	}

	token = token[7:]
	err := c.userService.Logout(token)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "登出失败"))
		return
	}

	userID := ctx.GetUint("userID")
	utils.Redis.Del(utils.Ctx, fmt.Sprintf("signkey:%d", userID))

	ctx.JSON(http.StatusOK, utils.Success("登出成功", nil))
}

func (c *AuthController) GetProfile(ctx *gin.Context) {
	userID := ctx.GetUint("userID")
	user, err := c.userService.GetUserByID(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, err.Error()))
		return
	}

	roleIds, _ := c.userService.GetUserRoleIds(user.ID)
	roleName := "用户"
	if len(roleIds) > 0 {
		role, err := c.roleService.GetRoleByID(uint(roleIds[0]))
		if err == nil {
			roleName = role.Name
		}
	}

	opCount, _ := c.logService.GetUserOperationCount(user.ID)
	onlineDays := max(int(math.Floor(time.Since(user.CreatedAt).Hours()/24)), 1)
	articleCount := c.articleService.GetArticleCountByAuthor(strconv.FormatUint(uint64(user.ID), 10))

	ctx.JSON(http.StatusOK, utils.Success("获取用户信息成功", gin.H{
		"user": gin.H{
			"id":             user.ID,
			"username":       user.Username,
			"email":          user.Email,
			"phone":          user.Mobile,
			"account":        user.Account,
			"roleName":       roleName,
			"avatar":         user.Avatar,
			"createdAt":      user.CreatedAt.Format("2006-01-02"),
			"onlineDays":     onlineDays,
			"articleCount":   articleCount,
			"operationCount": opCount,
			"bio":            user.Bio,
		},
	}))
}

func (c *AuthController) UpdateProfile(ctx *gin.Context) {
	userID := ctx.GetUint("userID")
	var req struct {
		Email  string `json:"email" binding:"required,email"`
		Phone  string `json:"phone"`
		Bio    string `json:"bio"`
		Avatar string `json:"avatar"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: "+err.Error()))
		return
	}

	err := c.userService.UpdateProfile(userID, req.Email, req.Phone, req.Bio, req.Avatar)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "更新失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("更新成功", nil))
}

func (c *AuthController) ChangePassword(ctx *gin.Context) {
	userID := ctx.GetUint("userID")
	var req struct {
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: "+err.Error()))
		return
	}

	settingsService := &services.SettingsService{}
	settings, err := settingsService.GetSettings()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取安全设置失败"))
		return
	}

	if len(req.NewPassword) < settings.MinPasswordLength {
		ctx.JSON(http.StatusOK, utils.Error(1, fmt.Sprintf("密码长度不能少于%d位", settings.MinPasswordLength)))
		return
	}

	err = c.userService.ChangePassword(userID, req.OldPassword, req.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("密码修改成功", nil))
}
