package service

import (
	"errors"
	"fmt"

	"base/internal/models"
	"base/pkg/db"
	"base/pkg/utils"
)

type UserService struct{}

func (s UserService) Create(u *models.User) error {
	var count int64
	if err := db.DB.Model(&models.User{}).
		Where("tenant_id = ? AND username = ?", u.TenantID, u.Username).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该租户下用户名已存在")
	}

	if u.Password == "" {
		u.Password = "123456"
	}
	// 密码最小长度来自「系统设置 → 安全策略」（默认 8 位）
	if minLen := (SettingsService{}).GetSecuritySettings().PwdMinLength; len([]rune(u.Password)) < minLen {
		return fmt.Errorf("密码长度不能少于 %d 位", minLen)
	}
	hash, err := utils.HashPassword(u.Password)
	if err != nil {
		return err
	}
	u.Password = hash
	return db.DB.Create(u).Error
}

func (s UserService) Update(u *models.User, tenantID uint64) error {
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
	db := db.DB.Model(u)
	if tenantID > 0 {
		db = db.Where("tenant_id = ?", tenantID)
	}
	return db.Updates(updates).Error
}

func (s UserService) Delete(id uint64, tenantID uint64) error {
	db := db.DB
	if tenantID > 0 {
		db = db.Where("tenant_id = ?", tenantID)
	}
	return db.Delete(&models.User{BaseModel: models.BaseModel{ID: id}}).Error
}

func (s UserService) GetByID(id uint64, tenantID uint64) (*models.User, error) {
	var u models.User
	db := db.DB.Preload("Roles")
	if tenantID > 0 {
		db = db.Where("tenant_id = ?", tenantID)
	}
	err := db.First(&u, id).Error
	return &u, err
}

func (s UserService) GetByUsername(username string) (*models.User, error) {
	var u models.User
	err := db.DB.Where("username = ?", username).First(&u).Error
	return &u, err
}

// List 分页查询用户。tenantID 为当前登录用户所属租户：
//   - 普通租户用户（tenantID > 0）只能看到本租户；
//   - 平台超管（tenantID == 0）不传 filterTenantID 时查看全部租户，传则只看指定租户。
func (s UserService) List(tenantID, filterTenantID uint64, page, size int, keyword string) ([]models.User, int64, error) {
	var list []models.User
	var total int64
	query := db.DB.Model(&models.User{})
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	} else if filterTenantID > 0 {
		query = query.Where("tenant_id = ?", filterTenantID)
	}
	if keyword != "" {
		query = query.Where("username LIKE ? OR real_name LIKE ? OR phone LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (s UserService) AssignRoles(userID uint64, roleIDs []uint64, tenantID uint64) error {
	var user models.User
	query := db.DB.Where("id = ?", userID)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&user).Error; err != nil {
		return err
	}
	var roles []models.Role
	if len(roleIDs) > 0 {
		// 角色必须与用户同租户，避免把其他租户的角色分配给本租户用户
		if err := db.DB.Where("id IN ? AND tenant_id = ?", roleIDs, user.TenantID).Find(&roles).Error; err != nil {
			return err
		}
		if len(roles) == 0 {
			return errors.New("所选角色不属于该用户所在租户")
		}
	}
	return db.DB.Model(&user).Association("Roles").Replace(roles)
}

func (s UserService) ResetPassword(userID uint64, tenantID uint64) error {
	var user models.User
	query := db.DB.Where("id = ?", userID)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&user).Error; err != nil {
		return err
	}
	hash, err := utils.HashPassword("123456")
	if err != nil {
		return err
	}
	return db.DB.Model(&user).Update("password", hash).Error
}

func (s UserService) ChangePassword(userID uint64, oldPwd, newPwd string) error {
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return err
	}
	if !utils.CheckPassword(oldPwd, user.Password) {
		return errors.New("旧密码错误")
	}
	// 密码最小长度来自「系统设置 → 安全策略」（默认 8 位）
	if minLen := (SettingsService{}).GetSecuritySettings().PwdMinLength; len([]rune(newPwd)) < minLen {
		return fmt.Errorf("新密码长度不能少于 %d 位", minLen)
	}
	hash, err := utils.HashPassword(newPwd)
	if err != nil {
		return err
	}
	return db.DB.Model(&user).Update("password", hash).Error
}
