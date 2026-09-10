package service

import (
	"errors"

	"application/internal/models"
	"application/pkg/db"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct{}

// --- Applicant management (申报人) ---

func (s *UserService) ListUsers(page, size int, keyword string) ([]models.User, int64, error) {
	var list []models.User
	var total int64
	query := db.DB.Model(&models.User{})
	if keyword != "" {
		query = query.Where("username LIKE ? OR real_name LIKE ? OR organization LIKE ? OR phone LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (s *UserService) CreateUser(u *models.User) error {
	if u.Username == "" || u.Password == "" || u.RealName == "" {
		return errors.New("请填写完整信息")
	}
	var count int64
	db.DB.Model(&models.User{}).Where("username = ?", u.Username).Count(&count)
	if count > 0 {
		return errors.New("用户名已存在")
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	u.Password = string(hashed)
	u.Status = 1
	return db.DB.Create(u).Error
}

func (s *UserService) UpdateUser(id uint64, updates map[string]interface{}) error {
	delete(updates, "id")
	delete(updates, "username")
	if pwd, ok := updates["password"].(string); ok && pwd != "" {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		updates["password"] = string(hashed)
	} else {
		delete(updates, "password")
	}
	return db.DB.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error
}

func (s *UserService) DeleteUser(id uint64) error {
	return db.DB.Delete(&models.User{}, id).Error
}

func (s *UserService) SetUserStatus(id uint64, status int) error {
	return db.DB.Model(&models.User{}).Where("id = ?", id).Update("status", status).Error
}

// --- Admin management (管理人 / 评审人) ---

func (s *UserService) ListAdmins(page, size int, keyword string) ([]models.Admin, int64, error) {
	var list []models.Admin
	var total int64
	query := db.DB.Model(&models.Admin{})
	if keyword != "" {
		query = query.Where("username LIKE ? OR real_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (s *UserService) CreateAdmin(a *models.Admin) error {
	if a.Username == "" || a.Password == "" || a.RealName == "" || a.RoleCode == "" {
		return errors.New("请填写完整信息")
	}
	var count int64
	db.DB.Model(&models.Admin{}).Where("username = ?", a.Username).Count(&count)
	if count > 0 {
		return errors.New("用户名已存在")
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte(a.Password), bcrypt.DefaultCost)
	a.Password = string(hashed)
	a.Status = 1
	return db.DB.Create(a).Error
}

func (s *UserService) UpdateAdmin(id uint64, updates map[string]interface{}) error {
	delete(updates, "id")
	delete(updates, "username")
	if pwd, ok := updates["password"].(string); ok && pwd != "" {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		updates["password"] = string(hashed)
	} else {
		delete(updates, "password")
	}
	return db.DB.Model(&models.Admin{}).Where("id = ?", id).Updates(updates).Error
}

func (s *UserService) DeleteAdmin(id uint64) error {
	var count int64
	db.DB.Model(&models.ReviewAssignment{}).Where("reviewer_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("该账号存在评审任务，无法删除")
	}
	return db.DB.Delete(&models.Admin{}, id).Error
}

func (s *UserService) ListRoles() ([]models.Role, error) {
	var list []models.Role
	err := db.DB.Order("id ASC").Find(&list).Error
	return list, err
}
