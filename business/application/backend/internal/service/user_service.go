package service

import (
	"errors"

	"application/internal/models"
	"application/pkg/db"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct{}

// UserRequest is the creation payload for an applicant. It exists because the
// model hides Password (`json:"-"`), so binding a request straight into the
// model silently dropped the password and every create failed.
type UserRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	RealName     string `json:"realName"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	IDCard       string `json:"idCard"`
	Organization string `json:"organization"`
	Position     string `json:"position"`
}

func (r UserRequest) ToModel() models.User {
	return models.User{
		Username:     r.Username,
		Password:     r.Password,
		RealName:     r.RealName,
		Phone:        r.Phone,
		Email:        r.Email,
		IDCard:       r.IDCard,
		Organization: r.Organization,
		Position:     r.Position,
	}
}

// AdminRequest is the creation payload for a manager account (see UserRequest).
type AdminRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	RealName string `json:"realName"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	RoleCode string `json:"roleCode"`
}

func (r AdminRequest) ToModel() models.Admin {
	return models.Admin{
		Username: r.Username,
		Password: r.Password,
		RealName: r.RealName,
		Phone:    r.Phone,
		Email:    r.Email,
		RoleCode: r.RoleCode,
	}
}

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
	// 评审人账号只能从「专家库」新增：两个入口都能建 reviewer 会产生没有专家
	// 档案、职责不明的重复账号。
	if a.RoleCode == models.RoleReviewer {
		return errors.New("评审人账号请通过「专家库」新增")
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
		// 从专家库新建的评审人不能改回评审人角色以外的用途，反之也不允许在
		// 账号管理里新造一个评审人（见 CreateAdmin）。已有的评审人账号仍可编辑
		// 资料，因为它的角色没有变化。
		if role == models.RoleReviewer {
			var target models.Admin
			if err := db.DB.First(&target, id).Error; err == nil && target.RoleCode != models.RoleReviewer {
				return errors.New("评审人账号请通过「专家库」新增")
			}
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
	// A 专家库 account owns a profile row next to its login; deleting only the
	// login would leave an orphan expert that can never sign in again. Both
	// entry points (账号管理 / 专家库) therefore remove the pair together.
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("admin_id = ?", id).Delete(&models.Expert{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Admin{}, id).Error
	})
}

func (s *UserService) ListRoles() ([]models.Role, error) {
	var list []models.Role
	err := db.DB.Order("id ASC").Find(&list).Error
	return list, err
}
