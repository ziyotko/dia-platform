package service

import (
	"errors"
	"time"

	"base/config"
	"base/internal/models"
	"base/pkg/db"
	"base/pkg/jwt"
	"base/pkg/utils"

	"github.com/sirupsen/logrus"
)

type AuthService struct{}

type LoginDTO struct {
	Username    string
	Password    string
	TenantCode  string
	CaptchaID   string
	CaptchaCode string
}

// Login 校验账号密码，成功时返回用户与「access + refresh」令牌对。
func (s AuthService) Login(dto LoginDTO) (*models.User, TokenPair, error) {
	captchaSvc := CaptchaService{}
	security := SettingsService{}.GetSecuritySettings()

	// 解析租户编码，为空则视为平台级（tenantID=0）
	var tenantID uint64
	if dto.TenantCode != "" {
		var tenant models.Tenant
		if err := db.DB.Where("code = ?", dto.TenantCode).First(&tenant).Error; err != nil {
			return nil, TokenPair{}, errors.New("租户不存在")
		}
		if tenant.Status != 1 {
			return nil, TokenPair{}, errors.New("租户已禁用")
		}
		tenantID = tenant.ID
	}

	// 检查账号是否因登录失败被锁定（按租户隔离）
	if security.LockEnabled {
		if err := captchaSvc.CheckAndLock(tenantID, dto.Username, security.MaxFailCount, security.LockDuration); err != nil {
			return nil, TokenPair{}, err
		}
	}

	// 验证码开关由系统设置（安全策略 → captchaEnabled）决定，关闭后不再校验
	if security.CaptchaEnabled && !captchaSvc.Verify(dto.CaptchaID, dto.CaptchaCode) {
		if security.LockEnabled {
			_, _ = captchaSvc.RecordLoginFail(tenantID, dto.Username, security.MaxFailCount, security.LockDuration)
		}
		return nil, TokenPair{}, errors.New("验证码错误")
	}

	var user models.User
	if err := db.DB.Where("username = ? AND tenant_id = ?", dto.Username, tenantID).First(&user).Error; err != nil {
		if security.LockEnabled {
			_, _ = captchaSvc.RecordLoginFail(tenantID, dto.Username, security.MaxFailCount, security.LockDuration)
		}
		return nil, TokenPair{}, errors.New("用户不存在")
	}
	if user.Status != 1 {
		if security.LockEnabled {
			_, _ = captchaSvc.RecordLoginFail(tenantID, dto.Username, security.MaxFailCount, security.LockDuration)
		}
		return nil, TokenPair{}, errors.New("账号已禁用")
	}
	if !utils.CheckPassword(dto.Password, user.Password) {
		if security.LockEnabled {
			_, _ = captchaSvc.RecordLoginFail(tenantID, dto.Username, security.MaxFailCount, security.LockDuration)
		}
		return nil, TokenPair{}, errors.New("密码错误")
	}

	// 登录成功，清除失败次数
	_ = captchaSvc.ClearLoginFail(tenantID, dto.Username)

	pair, err := s.IssueTokenPair(&user)
	if err != nil {
		return nil, TokenPair{}, err
	}
	return &user, pair, nil
}

// IssueTokenPair 为一个已通过校验的用户签发 access token + refresh token。
// access token 是 JWT（无状态、短生命周期，默认 8 小时）；refresh token 是随机串，
// 状态在 Redis（可轮换、可复用检测、登出即失效），用于免登录续期。
func (s AuthService) IssueTokenPair(user *models.User) (TokenPair, error) {
	access, err := jwt.GenerateToken(user.ID, user.Username, user.TenantID)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, err := (RefreshTokenService{}).Issue(user.ID)
	if err != nil {
		return TokenPair{}, err
	}
	hours := config.Cfg.JWT.ExpireHours
	if hours <= 0 {
		hours = 8
	}
	return TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(hours) * 3600,
	}, nil
}

// RefreshSession 用 refresh token 换一对新令牌（轮换 + 复用检测 + 宽限期重放）。
//
//   - 宽限期内用同一个 refresh token 重放 → 返回上次签发的那一对（幂等，容忍并发/重试）；
//   - 正常轮换 → 旧 refresh token 立即失效，签发新的；
//   - 已消费的 refresh token 再次出现（超过宽限期）→ 判定泄漏，吊销该用户全部会话；
//   - 用户被禁用/删除/改密（auth:user:revoked-before）→ 该用户所有 refresh token 同样失效。
func (s AuthService) RefreshSession(refreshToken string) (TokenPair, error) {
	refreshSvc := RefreshTokenService{}
	if pair, ok := refreshSvc.Replay(refreshToken); ok {
		return pair, nil
	}

	userID, err := refreshSvc.Load(refreshToken)
	if err != nil {
		// 已被消费过还来续期：视为凭证泄漏，直接吊销该用户全部会话（refresh 校验会带上 revoked-before）
		if uid, reuse := refreshSvc.ReuseDetected(refreshToken); reuse {
			if revokeErr := (TokenService{}).RevokeUserTokensBefore(uid, time.Now()); revokeErr != nil {
				logrus.WithError(revokeErr).Warn("复用检测后吊销用户会话失败")
			}
			logrus.Warnf("检测到 refresh token 复用，已吊销用户 %d 的全部会话", uid)
			return TokenPair{}, errRefreshReuse
		}
		return TokenPair{}, err
	}

	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return TokenPair{}, errRefreshInvalid
	}
	if user.Status != 1 {
		return TokenPair{}, errors.New("账号已禁用")
	}

	pair, err := s.IssueTokenPair(&user)
	if err != nil {
		return TokenPair{}, err
	}
	if err := refreshSvc.Consume(refreshToken, user.ID, pair); err != nil {
		logrus.WithError(err).Warn("记录 refresh token 轮换状态失败")
	}
	return pair, nil
}

// IssueAppTicket 为当前登录用户签发子应用一次性接入票据。
func (s AuthService) IssueAppTicket(userID uint64) (string, int, error) {
	ttl := AppTicketTTL
	if config.Cfg != nil && config.Cfg.Server.AppTicketTTLSeconds > 0 {
		ttl = time.Duration(config.Cfg.Server.AppTicketTTLSeconds) * time.Second
	}
	return (AppTicketService{}).Issue(userID, ttl)
}

// ExchangeAppTicket 用一次性票据换回子应用可用的会话（access token + 用户信息）。
// 只返回 access token，不下发 refresh token：子应用不能在后台无限续期，
// 需要新会话时由底座重新签发票据。
func (s AuthService) ExchangeAppTicket(ticket string) (*models.User, string, int64, error) {
	user, err := (AppTicketService{}).Consume(ticket)
	if err != nil {
		return nil, "", 0, err
	}
	access, err := jwt.GenerateToken(user.ID, user.Username, user.TenantID)
	if err != nil {
		return nil, "", 0, err
	}
	hours := config.Cfg.JWT.ExpireHours
	if hours <= 0 {
		hours = 8
	}
	return user, access, int64(hours) * 3600, nil
}

func (s AuthService) GetUserInfo(userID uint64) (*models.User, error) {
	var user models.User
	err := db.DB.Preload("Roles").First(&user, userID).Error
	return &user, err
}

func (s AuthService) GetUserPermissions(userID uint64) ([]string, error) {
	var user models.User
	if err := db.DB.Preload("Roles.Perms").First(&user, userID).Error; err != nil {
		return nil, err
	}
	permSet := make(map[string]struct{})
	for _, role := range user.Roles {
		for _, p := range role.Perms {
			if p.Status == 1 {
				permSet[p.Code] = struct{}{}
			}
		}
	}
	var list []string
	for code := range permSet {
		list = append(list, code)
	}
	return list, nil
}

// EnsureSuperAdmin 幂等地保证平台超级管理员存在（首次部署/被误删后可恢复）。
// 返回是否本次新建。平台超级管理员是底座唯一的初始账号，实现只保留在这里，
// 启动流程（main）与 /auth/init 接口共用，避免两份实现口径不一致。
func (s AuthService) EnsureSuperAdmin(password string) (bool, error) {
	var count int64
	if err := db.DB.Model(&models.User{}).
		Where("tenant_id = ? AND username = ?", models.PlatformTenantID, "admin").
		Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return false, nil
	}

	if password == "" {
		password = "admin123"
	}
	admin := models.User{
		TenantID: models.PlatformTenantID,
		Username: "admin",
		Password: password,
		RealName: "超级管理员",
		Status:   1,
		IsAdmin:  true,
	}
	// 密码策略（pwdMinLength）与用户名唯一性统一由 UserService.Create 负责，避免这里再写一份
	if err := (UserService{}).Create(&admin); err != nil {
		return false, err
	}
	return true, nil
}

// InitSuperAdmin 首次部署时创建平台超级管理员（/auth/init）。
// 已存在时返回错误而不会重置密码：重置请由管理员登录后在「用户管理 → 重置密码」操作。
func (s AuthService) InitSuperAdmin(password string) error {
	created, err := s.EnsureSuperAdmin(password)
	if err != nil {
		return err
	}
	if !created {
		return errors.New("超级管理员已存在")
	}
	return nil
}
