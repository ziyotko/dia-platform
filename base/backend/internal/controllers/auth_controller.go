package controllers

import (
	"time"

	"base/config"
	"base/internal/models"
	"base/internal/service"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AuthController struct {
	authService service.AuthService
}

type LoginReq struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	TenantCode  string `json:"tenantCode"`
	CaptchaID   string `json:"captcha_id"`
	CaptchaCode string `json:"captcha_code"`
}

type LoginResp struct {
	// Token access token（JWT）；RefreshToken 用于免登录续期（不透明串，仅在响应体返回）
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	User         *models.User `json:"user"`
}

func (ctl *AuthController) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	user, pair, err := ctl.authService.Login(service.LoginDTO{
		Username:    req.Username,
		Password:    req.Password,
		TenantCode:  req.TenantCode,
		CaptchaID:   req.CaptchaID,
		CaptchaCode: req.CaptchaCode,
	})

	// 提取需要的数据，避免在 goroutine 中访问 gin.Context；
	// 同步写库：异步 goroutine 在进程退出/重启时会丢失登录日志
	ip := c.ClientIP()
	agent := c.Request.UserAgent()
	ctl.recordLoginLog(ip, agent, req.Username, user, err)

	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, LoginResp{Token: pair.AccessToken, RefreshToken: pair.RefreshToken, ExpiresIn: pair.ExpiresIn, User: user})
}

// Refresh 用 refresh token 换发新的 access + refresh（轮换）。
// 公开接口（不能依赖已过期的 access token），因此按 IP 限流；凭证在请求体里，不会出现在 URL/日志。
func (ctl *AuthController) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		response.FailWithCode(c, response.CodeUnauthorized, "登录状态已失效，请重新登录")
		return
	}
	pair, err := ctl.authService.RefreshSession(req.RefreshToken)
	if err != nil {
		response.FailWithCode(c, response.CodeUnauthorized, err.Error())
		return
	}
	response.Ok(c, pair)
}

// AppTicket 为当前登录用户签发子应用一次性接入票据（用于 iframe/新窗口打开子应用）。
// 子应用拿 base_ticket 调 POST /auth/app-ticket/exchange 换回 access token + 用户信息，
// 避免把长期凭证直接放在 URL query 里（浏览器历史、Referer、网关日志都会留痕）。
func (ctl *AuthController) AppTicket(c *gin.Context) {
	ticket, expiresIn, err := ctl.authService.IssueAppTicket(c.GetUint64("userID"))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, gin.H{"ticket": ticket, "expires_in": expiresIn})
}

// ExchangeAppTicket 子应用用一次性票据换回会话（公开接口 + 按 IP 限流，票据本身一次性）。
func (ctl *AuthController) ExchangeAppTicket(c *gin.Context) {
	var req struct {
		Ticket string `json:"ticket"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Ticket == "" {
		response.FailWithCode(c, response.CodeBadRequest, "缺少票据")
		return
	}
	user, access, expiresIn, err := ctl.authService.ExchangeAppTicket(req.Ticket)
	if err != nil {
		response.FailWithCode(c, response.CodeUnauthorized, err.Error())
		return
	}
	response.Ok(c, gin.H{
		"token":      access,
		"expires_in": expiresIn,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"realName": user.RealName,
			"tenantId": user.TenantID,
			"isAdmin":  user.IsAdmin,
		},
	})
}

func (ctl *AuthController) recordLoginLog(ip, agent, username string, user *models.User, loginErr error) {
	logSvc := service.LoginLogService{}
	log := models.LoginLog{
		Username: username,
		IP:       ip,
		Agent:    agent,
		Status:   1,
	}
	if user != nil {
		log.TenantID = user.TenantID
		log.UserID = user.ID
	}
	if loginErr != nil {
		log.Status = 0
		log.Message = loginErr.Error()
	} else {
		log.Message = "登录成功"
	}
	_ = logSvc.Create(&log)
}

func (ctl *AuthController) Captcha(c *gin.Context) {
	id, b64s, err := service.CaptchaService{}.Generate()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	// 字段名与 portal / member / application 保持一致：captcha_id + captcha_img
	response.Ok(c, gin.H{
		"captcha_id":  id,
		"captcha_img": b64s,
	})
}

// Logout 登出：把当前 access token 加入 Redis 黑名单，并作废请求体里的 refresh token。
// 前端在调用后清空本地会话（即使本接口失败也不阻断前端登出）。
func (ctl *AuthController) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.ShouldBindJSON(&req)

	expiresAt := time.Now().Add(time.Duration(config.Cfg.JWT.ExpireHours) * time.Hour)
	if v, ok := c.Get("tokenExp"); ok {
		if t, ok := v.(time.Time); ok {
			expiresAt = t
		}
	}
	if err := (service.TokenService{}).RevokeToken(c.GetString("tokenID"), expiresAt); err != nil {
		response.Fail(c, err.Error())
		return
	}
	// refresh token 也必须作废：否则登出后仍能用它换回新 token
	if err := (service.RefreshTokenService{}).Revoke(req.RefreshToken, c.GetUint64("userID")); err != nil {
		logrus.WithError(err).Warn("登出时作废 refresh token 失败")
	}
	response.OkWithMessage(c, "已退出登录", nil)
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
	// 改密后让该用户此前签发的 token 全部失效，避免旧 token 继续可用
	if err := (service.TokenService{}).RevokeUserTokensBefore(userID, time.Now()); err != nil {
		logrus.WithError(err).Warn("改密后吊销旧 token 失败")
	}
	response.OkWithMessage(c, "密码修改成功，请重新登录", nil)
}
