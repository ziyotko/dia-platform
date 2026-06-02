package services

import (
	"errors"
	"strconv"
	"strings"

	"server/models"
	"server/utils"
)

type UserService struct{}

type UserListResult struct {
	Total int64         `json:"total"`
	List  []models.User `json:"list"`
}

func (s *UserService) Login(email, account, mobile, password, captchaID, captchaCode string) (*models.User, string, error) {
	if !utils.VerifyCaptcha(captchaID, captchaCode) {
		return nil, "", errors.New("验证码错误")
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
		return nil, "", errors.New("用户不存在")
	}

	if err != nil {
		return nil, "", errors.New("用户不存在")
	}

	if user.Status != 1 {
		return nil, "", errors.New("用户已禁用")
	}

	if !user.ComparePassword(password) {
		return nil, "", errors.New("密码错误")
	}

	token, err := utils.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, "", err
	}

	return &user, token, nil
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

func (s *UserService) CreateUser(username, nickname, account, email, password, phone string, status int, roleIds []int) error {
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

	return utils.DB.Create(user).Error
}

func (s *UserService) UpdateUser(id uint, username, nickname, account, email, password, phone string, status int, roleIds []int) error {
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

	return utils.DB.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error
}

func (s *UserService) DeleteUser(id uint) error {
	return utils.DB.Delete(&models.User{}, id).Error
}

func (s *UserService) UpdateUserStatus(id uint, status int) error {
	return utils.DB.Model(&models.User{}).Where("id = ?", id).Update("status", status).Error
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
