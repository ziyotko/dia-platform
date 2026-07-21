package service

import (
	"errors"

	"base/internal/models"
	"base/pkg/db"
	"base/pkg/utils"
)

type UserService struct{}

func (s UserService) Create(u *models.User) error {
	if u.Password == "" {
		u.Password = "123456"
	}
	hash, err := utils.HashPassword(u.Password)
	if err != nil {
		return err
	}
	u.Password = hash
	return db.DB.Create(u).Error
}

func (s UserService) Update(u *models.User) error {
	updates := map[string]interface{}{
		"real_name": u.RealName,
		"phone":     u.Phone,
		"email":     u.Email,
		"avatar":    u.Avatar,
		"status":    u.Status,
		"is_admin":  u.IsAdmin,
	}
	if u.Password != "" {
		hash, err := utils.HashPassword(u.Password)
		if err != nil {
			return err
		}
		updates["password"] = hash
	}
	return db.DB.Model(u).Updates(updates).Error
}

func (s UserService) Delete(id uint64) error {
	return db.DB.Delete(&models.User{BaseModel: models.BaseModel{ID: id}}).Error
}

func (s UserService) GetByID(id uint64) (*models.User, error) {
	var u models.User
	err := db.DB.Preload("Roles").First(&u, id).Error
	return &u, err
}

func (s UserService) GetByUsername(username string) (*models.User, error) {
	var u models.User
	err := db.DB.Where("username = ?", username).First(&u).Error
	return &u, err
}

func (s UserService) List(tenantID uint64, page, size int, keyword string) ([]models.User, int64, error) {
	var list []models.User
	var total int64
	query := db.DB.Model(&models.User{}).Where("tenant_id = ?", tenantID)
	if keyword != "" {
		query = query.Where("username LIKE ? OR real_name LIKE ? OR phone LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (s UserService) AssignRoles(userID uint64, roleIDs []uint64) error {
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return err
	}
	var roles []models.Role
	if len(roleIDs) > 0 {
		if err := db.DB.Find(&roles, roleIDs).Error; err != nil {
			return err
		}
	}
	return db.DB.Model(&user).Association("Roles").Replace(roles)
}

func (s UserService) ResetPassword(userID uint64) error {
	hash, err := utils.HashPassword("123456")
	if err != nil {
		return err
	}
	return db.DB.Model(&models.User{BaseModel: models.BaseModel{ID: userID}}).Update("password", hash).Error
}

func (s UserService) ChangePassword(userID uint64, oldPwd, newPwd string) error {
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return err
	}
	if !utils.CheckPassword(oldPwd, user.Password) {
		return errors.New("旧密码错误")
	}
	hash, err := utils.HashPassword(newPwd)
	if err != nil {
		return err
	}
	return db.DB.Model(&user).Update("password", hash).Error
}
