package service

import (
	"base/internal/models"
	"base/pkg/db"
	"base/pkg/jwt"
	"base/pkg/utils"
	"errors"
)

type AuthService struct{}

func (s AuthService) Login(username, password string) (*models.User, string, error) {
	var user models.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, "", errors.New("用户不存在")
	}
	if user.Status != 1 {
		return nil, "", errors.New("账号已禁用")
	}
	if !utils.CheckPassword(password, user.Password) {
		return nil, "", errors.New("密码错误")
	}
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

func (s AuthService) InitSuperAdmin(password string) error {
	var count int64
	db.DB.Model(&models.User{}).Where("tenant_id = ?", 0).Count(&count)
	if count > 0 {
		return errors.New("超级管理员已存在")
	}
	hash, _ := utils.HashPassword(password)
	admin := models.User{
		TenantID: 0,
		Username: "admin",
		Password: hash,
		RealName: "超级管理员",
		Status:   1,
		IsAdmin:  true,
	}
	return db.DB.Create(&admin).Error
}
