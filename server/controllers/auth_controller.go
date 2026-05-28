package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"server/services"
	"server/utils"
)

type AuthController struct {
	userService *services.UserService
}

func NewAuthController() *AuthController {
	return &AuthController{
		userService: &services.UserService{},
	}
}

func (c *AuthController) GetCaptcha(ctx *gin.Context) {
	id, b64s, err := utils.GenerateCaptcha()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "生成验证码失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"captcha_id": id, "captcha_img": b64s})
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
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	user, token, err := c.userService.Login(req.Email, req.Account, req.Mobile, req.Password, req.CaptchaID, req.CaptchaCode)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

func (c *AuthController) Logout(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "未提供token"})
		return
	}

	token = token[7:]
	err := c.userService.Logout(token)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "登出失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "登出成功"})
}

func (c *AuthController) GetProfile(ctx *gin.Context) {
	userID := ctx.GetUint("userID")
	user, err := c.userService.GetUserByID(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"user": user})
}
