package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"server/config"
	"server/models"
	"server/utils"
)

type UserService struct{}

type UserListResult struct {
	Total int64         `json:"total"`
	List  []models.User `json:"list"`
}

func (s *UserService) Login(email, account, mobile, password, captchaID, captchaCode string) (*models.User, string, string, error) {
	if !utils.VerifyCaptcha(captchaID, captchaCode) {
		return nil, "", "", errors.New("验证码错误")
	}

	var user models.User
	var err error

	if email != "" {
		err = utils.DB.Where("email = ?", email).First(&user).Error
	} else if account != "" {
		err = utils.DB.Where("account = ?", account).First(&user).Error
	} else if mobile != "" {
		err = utils.DB.Where("mobile = ?", mobile).First(&user).Error
	} else {
		return nil, "", "", errors.New("用户不存在")
	}

	if err != nil {
		return nil, "", "", errors.New("用户不存在")
	}

	if user.Status != 1 {
		return nil, "", "", errors.New("用户已禁用")
	}

	settingsService := SettingsService{}
	settings, err := settingsService.GetSettings()
	if err != nil {
		return nil, "", "", errors.New("获取系统设置失败")
	}

	if settings.LockEnabled {
		if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
			remaining := int(time.Until(*user.LockedUntil).Minutes()) + 1
			return nil, "", "", fmt.Errorf("登录失败次数过多，请 %d 分钟后重试", remaining)
		}

		if !user.ComparePassword(password) {
			user.LoginFailCount++
			if user.LoginFailCount >= settings.MaxFailCount {
				lockUntil := time.Now().Add(time.Duration(settings.LockDuration) * time.Minute)
				user.LockedUntil = &lockUntil
				user.LoginFailCount = 0
				utils.DB.Model(&user).Updates(map[string]interface{}{
					"login_fail_count": 0,
					"locked_until":     lockUntil,
				})
				return nil, "", "", fmt.Errorf("登录失败次数过多，账号已锁定 %d 分钟", settings.LockDuration)
			}
			utils.DB.Model(&user).Update("login_fail_count", user.LoginFailCount)
			return nil, "", "", errors.New("密码错误")
		}

		if user.LoginFailCount > 0 || user.LockedUntil != nil {
			utils.DB.Model(&user).Updates(map[string]interface{}{
				"login_fail_count": 0,
				"locked_until":     nil,
			})
		}
	} else {
		if !user.ComparePassword(password) {
			return nil, "", "", errors.New("密码错误")
		}
	}

	expiresHour := 0
	if settings != nil && settings.TokenExpire > 0 {
		expiresHour = settings.TokenExpire
	}
	token, err := utils.GenerateToken(user.ID, user.Email, expiresHour)
	if err != nil {
		return nil, "", "", err
	}

	signKey, err := utils.GenerateSignKey()
	if err != nil {
		return nil, "", "", err
	}

	ttl := time.Duration(config.AppConfig.JWT.ExpiresHour) * time.Hour
	if expiresHour > 0 {
		ttl = time.Duration(expiresHour) * time.Hour
	}
	signKeyKey := fmt.Sprintf("signkey:%d", user.ID)
	if err := utils.Redis.Set(utils.Ctx, signKeyKey, signKey, ttl).Err(); err != nil {
		return nil, "", "", err
	}

	return &user, token, signKey, nil
}

func (s *UserService) Logout(token string) error {
	claims, err := utils.ParseToken(token)
	if err != nil {
		return err
	}

	err = utils.Redis.Set(utils.Ctx, "blacklist:"+claims.ID, token, 0).Err()
	return err
}

func (s *UserService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := utils.DB.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	return &user, nil
}

func (s *UserService) GetUserList(page, pageSize int, username, account string, status *int) (*UserListResult, error) {
	var users []models.User
	var total int64

	query := utils.DB.Model(&models.User{})

	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if account != "" {
		query = query.Where("account LIKE ?", "%"+account+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	err = query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&users).Error
	if err != nil {
		return nil, err
	}

	return &UserListResult{
		Total: total,
		List:  users,
	}, nil
}

func (s *UserService) CreateUser(username, nickname, account, email, password, phone string, status int, roleIds []int, orgId uint) error {
	if password == "" {
		password = "123456"
	}

	user := &models.User{
		Username: username,
		Nickname: nickname,
		Account:  account,
		Email:    email,
		Password: password,
		Mobile:   phone,
		Status:   status,
	}

	if len(roleIds) > 0 {
		roleIdsStr := make([]string, len(roleIds))
		for i, id := range roleIds {
			roleIdsStr[i] = strconv.Itoa(id)
		}
		user.RoleIds = strings.Join(roleIdsStr, ",")
	}

	if err := utils.DB.Create(user).Error; err != nil {
		return err
	}

	if orgId > 0 {
		orgService := OrganizationService{}
		if err := orgService.AddUserToOrganization(orgId, user.ID); err != nil {
			return err
		}
	}

	return nil
}

func (s *UserService) UpdateUser(id uint, username, nickname, account, email, password, phone string, status int, roleIds []int, orgId uint) error {
	updates := map[string]interface{}{
		"username": username,
		"nickname": nickname,
		"account":  account,
		"email":    email,
		"mobile":   phone,
		"status":   status,
	}

	if password != "" {
		updates["password"] = password
	}

	if len(roleIds) > 0 {
		roleIdsStr := make([]string, len(roleIds))
		for i, id := range roleIds {
			roleIdsStr[i] = strconv.Itoa(id)
		}
		updates["role_ids"] = strings.Join(roleIdsStr, ",")
	}

	if err := utils.DB.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return err
	}

	orgService := OrganizationService{}
	oldOrg, _ := orgService.GetOrganizationByUserId(id)
	oldOrgId := uint(0)
	if oldOrg != nil {
		oldOrgId = oldOrg.ID
	}

	if oldOrgId != orgId {
		if orgId > 0 {
			if err := orgService.AddUserToOrganization(orgId, id); err != nil {
				return err
			}
		}
		if oldOrgId > 0 {
			if err := orgService.RemoveUserFromOrganization(oldOrgId, id); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *UserService) DeleteUser(id uint) error {
	orgService := OrganizationService{}
	if err := orgService.RemoveUserFromAllOrganizations(id); err != nil {
		return err
	}
	return utils.DB.Unscoped().Delete(&models.User{}, id).Error
}

func (s *UserService) UpdateUserStatus(id uint, status int) error {
	return utils.DB.Model(&models.User{}).Where("id = ?", id).Update("status", status).Error
}

func (s *UserService) UpdateProfile(id uint, nickname, email, phone, bio, avatar string) error {
	return utils.DB.Model(&models.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"nickname": nickname,
		"email":    email,
		"mobile":   phone,
		"bio":      bio,
		"avatar":   avatar,
	}).Error
}

func (s *UserService) ChangePassword(id uint, oldPassword, newPassword string) error {
	user, err := s.GetUserByID(id)
	if err != nil {
		return err
	}
	if !user.ComparePassword(oldPassword) {
		return errors.New("原密码错误")
	}
	return utils.DB.Model(&models.User{}).Where("id = ?", id).Update("password", newPassword).Error
}

func (s *UserService) GetUserRoleIds(userId uint) ([]int, error) {
	var user models.User
	if err := utils.DB.First(&user, userId).Error; err != nil {
		return nil, err
	}

	if user.RoleIds == "" {
		return []int{}, nil
	}

	roleIdStrs := strings.Split(user.RoleIds, ",")
	roleIds := make([]int, 0, len(roleIdStrs))
	for _, idStr := range roleIdStrs {
		if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
			roleIds = append(roleIds, id)
		}
	}

	return roleIds, nil
}
