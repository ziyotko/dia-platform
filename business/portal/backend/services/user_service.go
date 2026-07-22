package services

import (
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"server/config"
	"server/models"
	"server/utils"

	"github.com/xuri/excelize/v2"
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
				utils.DB.Model(&user).Updates(map[string]any{
					"login_fail_count": 0,
					"locked_until":     lockUntil,
				})
				return nil, "", "", fmt.Errorf("登录失败次数过多，账号已锁定 %d 分钟", settings.LockDuration)
			}
			utils.DB.Model(&user).Update("login_fail_count", user.LoginFailCount)
			return nil, "", "", errors.New("密码错误")
		}

		if user.LoginFailCount > 0 || user.LockedUntil != nil {
			utils.DB.Model(&user).Updates(map[string]any{
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

func (s *UserService) CreateUser(username, nickname, account, email, password, phone string, status int, roleIds []int, orgIds []uint) error {
	if password == "" {
		password = "Abcd@1234"
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

	if len(orgIds) > 0 {
		orgService := OrganizationService{}
		for _, orgId := range orgIds {
			if orgId > 0 {
				if err := orgService.AddUserToOrganization(orgId, user.ID); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

type ImportUserResult struct {
	SuccessCount int      `json:"successCount"`
	FailCount    int      `json:"failCount"`
	FailDetails  []string `json:"failDetails"`
}

func (s *UserService) ImportUsers(file multipart.File, fileSize int64) (*ImportUserResult, error) {
	f, err := excelize.OpenReader(file, excelize.Options{})
	if err != nil {
		return nil, fmt.Errorf("读取 Excel 失败: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, errors.New("Excel 工作表为空")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取工作表失败: %w", err)
	}

	if len(rows) < 2 {
		return nil, errors.New("Excel 数据行数不足")
	}

	result := &ImportUserResult{
		SuccessCount: 0,
		FailCount:    0,
		FailDetails:  make([]string, 0),
	}

	for i, row := range rows[1:] {
		lineNum := i + 2
		if len(row) < 5 {
			result.FailCount++
			result.FailDetails = append(result.FailDetails, fmt.Sprintf("第 %d 行: 字段数量不足", lineNum))
			continue
		}

		username := strings.TrimSpace(row[0])
		account := strings.TrimSpace(row[1])
		nickname := strings.TrimSpace(row[2])
		email := strings.TrimSpace(row[3])
		phone := strings.TrimSpace(row[4])

		if username == "" || account == "" || nickname == "" || email == "" || phone == "" {
			result.FailCount++
			result.FailDetails = append(result.FailDetails, fmt.Sprintf("第 %d 行: 存在空字段", lineNum))
			continue
		}

		if err := s.CreateUser(username, nickname, account, email, "Abcd@1234", phone, 1, []int{6}, nil); err != nil {
			result.FailCount++
			result.FailDetails = append(result.FailDetails, fmt.Sprintf("第 %d 行 (%s): %s", lineNum, account, err.Error()))
			continue
		}

		result.SuccessCount++
	}

	return result, nil
}

func (s *UserService) UpdateUser(id uint, username, nickname, account, email, password, phone string, status int, roleIds []int, orgIds []uint) error {
	updates := map[string]any{
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
	oldOrgs, _ := orgService.GetOrganizationsByUserId(id)
	oldOrgIdMap := make(map[uint]bool)
	for _, org := range oldOrgs {
		oldOrgIdMap[org.ID] = true
	}
	newOrgIdMap := make(map[uint]bool)
	for _, orgId := range orgIds {
		if orgId > 0 {
			newOrgIdMap[orgId] = true
		}
	}

	for _, org := range oldOrgs {
		if !newOrgIdMap[org.ID] {
			if err := orgService.RemoveUserFromOrganization(org.ID, id); err != nil {
				return err
			}
		}
	}
	for _, orgId := range orgIds {
		if orgId > 0 && !oldOrgIdMap[orgId] {
			if err := orgService.AddUserToOrganization(orgId, id); err != nil {
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
	return utils.DB.Model(&models.User{}).Where("id = ?", id).Updates(map[string]any{
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
	hashedPassword := utils.SM3HashPassword(newPassword)
	return utils.DB.Model(&models.User{}).Where("id = ?", id).Update("password", hashedPassword).Error
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
