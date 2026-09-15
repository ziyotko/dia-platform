package service

import (
	"errors"

	"base/internal/models"
	"base/pkg/db"
	"base/pkg/jwt"
	"base/pkg/utils"

	"gorm.io/gorm"
)

type AuthService struct{}

type LoginDTO struct {
	Username    string
	Password    string
	TenantCode  string
	CaptchaID   string
	CaptchaCode string
}

func (s AuthService) Login(dto LoginDTO) (*models.User, string, error) {
	captchaSvc := CaptchaService{}
	security := SettingsService{}.GetSecuritySettings()

	// 解析租户编码，为空则视为平台级（tenantID=0）
	var tenantID uint64
	if dto.TenantCode != "" {
		var tenant models.Tenant
		if err := db.DB.Where("code = ?", dto.TenantCode).First(&tenant).Error; err != nil {
			return nil, "", errors.New("租户不存在")
		}
		if tenant.Status != 1 {
			return nil, "", errors.New("租户已禁用")
		}
		tenantID = tenant.ID
	}

	// 检查账号是否因登录失败被锁定（按租户隔离）
	if security.LockEnabled {
		if err := captchaSvc.CheckAndLock(tenantID, dto.Username, security.MaxFailCount, security.LockDuration); err != nil {
			return nil, "", err
		}
	}

	// 验证码开关由系统设置（安全策略 → captchaEnabled）决定，关闭后不再校验
	if security.CaptchaEnabled && !captchaSvc.Verify(dto.CaptchaID, dto.CaptchaCode) {
		if security.LockEnabled {
			_, _ = captchaSvc.RecordLoginFail(tenantID, dto.Username, security.MaxFailCount, security.LockDuration)
		}
		return nil, "", errors.New("验证码错误")
	}

	var user models.User
	if err := db.DB.Where("username = ? AND tenant_id = ?", dto.Username, tenantID).First(&user).Error; err != nil {
		if security.LockEnabled {
			_, _ = captchaSvc.RecordLoginFail(tenantID, dto.Username, security.MaxFailCount, security.LockDuration)
		}
		return nil, "", errors.New("用户不存在")
	}
	if user.Status != 1 {
		if security.LockEnabled {
			_, _ = captchaSvc.RecordLoginFail(tenantID, dto.Username, security.MaxFailCount, security.LockDuration)
		}
		return nil, "", errors.New("账号已禁用")
	}
	if !utils.CheckPassword(dto.Password, user.Password) {
		if security.LockEnabled {
			_, _ = captchaSvc.RecordLoginFail(tenantID, dto.Username, security.MaxFailCount, security.LockDuration)
		}
		return nil, "", errors.New("密码错误")
	}

	// 登录成功，清除失败次数
	_ = captchaSvc.ClearLoginFail(tenantID, dto.Username)

	token, err := jwt.GenerateToken(user.ID, user.Username, user.TenantID)
	if err != nil {
		return nil, "", err
	}
	return &user, token, nil
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
	var existing models.User
	err := db.DB.Unscoped().
		Where("tenant_id = ? AND username = ?", models.PlatformTenantID, "admin").
		First(&existing).Error
	switch {
	case err == nil:
		if !existing.DeletedAt.Valid {
			return false, nil
		}
		// 已被软删除：唯一索引 (tenant_id, username) 仍占位，继续 Create 会撞唯一键，
		// 交给 UserService.Create 走恢复分支（下面保留 username=admin 即会命中）。
	case errors.Is(err, gorm.ErrRecordNotFound):
	default:
		return false, err
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
