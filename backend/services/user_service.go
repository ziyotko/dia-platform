package services

import (
	"errors"
	"server/models"
	"server/utils"
)

type UserService struct{}

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
