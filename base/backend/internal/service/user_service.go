package service

import (
	"errors"
	"fmt"
	"strings"

	"base/internal/models"
	"base/pkg/db"
	"base/pkg/utils"

	"gorm.io/gorm"
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

	// 初始密码由管理员显式指定：不再隐式默认 123456
	// （6 位无法通过「安全策略 → 密码最小长度」校验，默认密码必然报错）
	if u.Password == "" {
		return errors.New("请填写初始密码")
	}
	if err := validatePassword(u.Password); err != nil {
		return err
	}
	if err := validateOrganization(u.OrganizationID, u.TenantID); err != nil {
		return err
	}
	hash, err := utils.HashPassword(u.Password)
	if err != nil {
		return err
	}
	u.Password = hash
	return db.DB.Create(u).Error
}

// IsAdmin 判断用户是否为租户管理员（base_user.is_admin）。
// 接口权限中间件与子应用代理入口共用，避免两处各写一份查询。
func (s UserService) IsAdmin(userID uint64) bool {
	var count int64
	if err := db.DB.Model(&models.User{}).
		Where("id = ? AND is_admin = ?", userID, true).
		Count(&count).Error; err != nil {
		return false
	}
	return count > 0
}

// validatePassword 按「系统设置 → 安全策略」的密码最小长度校验明文密码。
// 新增用户、重置密码、修改密码统一走这里，避免出现绕过策略的密码。
func validatePassword(pwd string) error {
	minLen := (SettingsService{}).GetSecuritySettings().PwdMinLength
	if len([]rune(pwd)) < minLen {
		return fmt.Errorf("密码长度不能少于 %d 位", minLen)
	}
	return nil
}

// validateOrganization 校验所属机构存在，且属于该用户所在租户（或为平台内置机构）。
// orgID = 0 表示不分配机构。
func validateOrganization(orgID, tenantID uint64) error {
	if orgID == 0 {
		return nil
	}
	query := db.DB.Model(&models.Organization{}).Where("id = ?", orgID)
	if tenantID > 0 {
		query = query.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}
	if err := ensureRecordExists(query, "所选机构不存在或不属于当前租户"); err != nil {
		return err
	}
	return nil
}

func (s UserService) Update(u *models.User, tenantID uint64) error {
	updates := map[string]interface{}{
		"real_name":       u.RealName,
		"phone":           u.Phone,
		"email":           u.Email,
		"avatar":          u.Avatar,
		"status":          u.Status,
		"is_admin":        u.IsAdmin,
		"organization_id": u.OrganizationID,
	}
	if err := validateOrganization(u.OrganizationID, tenantID); err != nil {
		return err
	}
	if u.Password != "" {
		hash, err := utils.HashPassword(u.Password)
		if err != nil {
			return err
		}
		updates["password"] = hash
	}
	check := db.DB.Model(&models.User{}).Where("id = ?", u.ID)
	query := db.DB.Model(u)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
		check = check.Where("tenant_id = ?", tenantID)
	}
	if err := ensureRecordExists(check, "用户不存在或不属于当前租户"); err != nil {
		return err
	}
	return query.Updates(updates).Error
}

// Delete 删除用户：不能删除自己；平台内置 admin 账号受保护；
// 同时清理角色与流程角色关联，避免残留脏关联数据。
func (s UserService) Delete(id, operatorID uint64, tenantID uint64) error {
	if id == operatorID {
		return errors.New("不能删除当前登录用户")
	}
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var user models.User
		query := tx.Where("id = ?", id)
		if tenantID > 0 {
			query = query.Where("tenant_id = ?", tenantID)
		}
		if err := query.First(&user).Error; err != nil {
			return errors.New("用户不存在或不属于当前租户")
		}
		if models.IsPlatformTenant(user.TenantID) && user.Username == "admin" {
			return errors.New("平台超级管理员账号不允许删除")
		}
		if err := tx.Exec("DELETE FROM base_user_role WHERE user_id = ?", user.ID).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM base_workflow_role_user WHERE user_id = ?", user.ID).Error; err != nil {
			return err
		}
		return tx.Delete(&user).Error
	})
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

func (s UserService) ResetPassword(userID uint64, newPwd string, tenantID uint64) error {
	if newPwd == "" {
		return errors.New("请填写新密码")
	}
	if err := validatePassword(newPwd); err != nil {
		return err
	}
	var user models.User
	query := db.DB.Where("id = ?", userID)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&user).Error; err != nil {
		return err
	}
	hash, err := utils.HashPassword(newPwd)
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
	if err := validatePassword(newPwd); err != nil {
		return fmt.Errorf("新密码%s", strings.TrimPrefix(err.Error(), "密码"))
	}
	hash, err := utils.HashPassword(newPwd)
	if err != nil {
		return err
	}
	return db.DB.Model(&user).Update("password", hash).Error
}
