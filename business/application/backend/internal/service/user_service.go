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
	clean := pickUpdates(updates, "real_name", "phone", "email", "id_card", "organization", "position", "password")
	if pwd, ok := clean["password"].(string); ok && pwd != "" {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		clean["password"] = string(hashed)
	} else {
		delete(clean, "password")
	}
	if len(clean) == 0 {
		return nil
	}
	return db.DB.Model(&models.User{}).Where("id = ?", id).Updates(clean).Error
}

// DeleteUser removes an applicant. Applicants that already own applications are
// kept so the submitted records never lose their owner.
func (s *UserService) DeleteUser(id uint64) error {
	var count int64
	db.DB.Model(&models.Application{}).Where("user_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("该申报人存在申报记录，无法删除")
	}
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
	if !validRole(a.RoleCode) {
		return errors.New("角色不存在")
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

// validRole keeps role_code inside the three roles the permission table knows
// about; anything else would log in with no permissions at all.
func validRole(code string) bool {
	switch code {
	case models.RoleSuperAdmin, models.RoleManager, models.RoleReviewer:
		return true
	}
	return false
}

func (s *UserService) UpdateAdmin(id uint64, updates map[string]interface{}) error {
	clean := pickUpdates(updates, "real_name", "phone", "email", "role_code", "status", "password")
	if raw, ok := clean["role_code"]; ok {
		role, _ := raw.(string)
		if !validRole(role) {
			return errors.New("角色不存在")
		}
	}
	if pwd, ok := clean["password"].(string); ok && pwd != "" {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		clean["password"] = string(hashed)
	} else {
		delete(clean, "password")
	}
	if len(clean) == 0 {
		return nil
	}
	return db.DB.Model(&models.Admin{}).Where("id = ?", id).Updates(clean).Error
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
