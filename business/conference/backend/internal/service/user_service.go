package service

import (
	"errors"

	"conference/internal/models"
	"conference/pkg/db"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct{}

func (s *UserService) Create(user *models.User) error {
	var count int64
	db.DB.Model(&models.User{}).Where("username = ?", user.Username).Count(&count)
	if count > 0 {
		return errors.New("用户名已存在")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}
	user.Password = string(hashed)
	return db.DB.Create(user).Error
}

func (s *UserService) Update(id uint64, updates map[string]interface{}) error {
	delete(updates, "password")
	delete(updates, "username")
	delete(updates, "id")
	return db.DB.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error
}

func (s *UserService) Delete(id uint64) error {
	return db.DB.Delete(&models.User{}, id).Error
}

func (s *UserService) GetByID(id uint64) (*models.User, error) {
	var user models.User
	if err := db.DB.First(&user, id).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	return &user, nil
}

func (s *UserService) List(keyword, branch, level string, page, size int) ([]models.User, int64, error) {
	var list []models.User
	var total int64
	query := db.DB.Model(&models.User{})
	if keyword != "" {
		query = query.Where("real_name LIKE ? OR username LIKE ? OR phone LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if branch != "" {
		query = query.Where("branch = ?", branch)
	}
	if level != "" {
		query = query.Where("member_level = ?", level)
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (s *UserService) SetValidity(id uint64, isValid bool) error {
	return db.DB.Model(&models.User{}).Where("id = ?", id).Update("is_valid", isValid).Error
}

// Admin management
func (s *UserService) CreateAdmin(admin *models.Admin) error {
	var count int64
	db.DB.Model(&models.Admin{}).Where("username = ?", admin.Username).Count(&count)
	if count > 0 {
		return errors.New("管理员用户名已存在")
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
	admin.Password = string(hashed)
	return db.DB.Create(admin).Error
}

func (s *UserService) ListAdmins(page, size int) ([]models.Admin, int64, error) {
	var list []models.Admin
	var total int64
	db.DB.Model(&models.Admin{}).Count(&total)
	err := db.DB.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (s *UserService) UpdateAdmin(id uint64, updates map[string]interface{}) error {
	delete(updates, "password")
	delete(updates, "username")
	delete(updates, "id")
	return db.DB.Model(&models.Admin{}).Where("id = ?", id).Updates(updates).Error
}

// Role management
func (s *UserService) ListRoles() ([]models.Role, error) {
	var roles []models.Role
	err := db.DB.Find(&roles).Error
	return roles, err
}
