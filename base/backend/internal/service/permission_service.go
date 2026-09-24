package service

import (
	"errors"
	"fmt"
	"strings"

	"base/internal/models"
	"base/pkg/db"
	"base/pkg/permmatch"

	"gorm.io/gorm"
)

type PermissionService struct{}

func (s PermissionService) Create(p *models.Permission) error {
	if err := validatePermissionFields(p); err != nil {
		return err
	}
	return db.DB.Create(p).Error
}

// validatePermissionFields 校验权限点字段长度（对应 base_permission 的定长列）。
func validatePermissionFields(p *models.Permission) error {
	return validateLengths(
		fieldLen{"应用编码", p.AppCode, 64},
		fieldLen{"权限编码", p.Code, 128},
		fieldLen{"权限名称", p.Name, 128},
		fieldLen{"类型", p.Type, 32},
		fieldLen{"路径", p.Path, 256},
		fieldLen{"HTTP 方法", p.Method, 16},
	)
}

func (s PermissionService) Update(p *models.Permission) error {
	if err := validatePermissionFields(p); err != nil {
		return err
	}
	if err := ensureRecordExists(db.DB.Model(&models.Permission{}).Where("id = ?", p.ID), "权限不存在"); err != nil {
		return err
	}
	return db.DB.Model(p).Updates(map[string]interface{}{
		"app_code":  p.AppCode,
		"code":      p.Code,
		"name":      p.Name,
		"type":      p.Type,
		"parent_id": p.ParentID,
		"path":      p.Path,
		"method":    p.Method,
		"status":    p.Status,
	}).Error
}

// Delete 删除权限：存在子权限时拒绝删除；同时解除与角色的关联。
func (s PermissionService) Delete(id uint64) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var perm models.Permission
		if err := tx.Where("id = ?", id).First(&perm).Error; err != nil {
			return errors.New("权限不存在")
		}
		var children int64
		if err := tx.Model(&models.Permission{}).Where("parent_id = ?", perm.ID).Count(&children).Error; err != nil {
			return err
		}
		if children > 0 {
			return fmt.Errorf("该权限下还有 %d 个子权限，请先删除子权限", children)
		}
		if err := tx.Exec("DELETE FROM base_role_permission WHERE permission_id = ?", perm.ID).Error; err != nil {
			return err
		}
		return tx.Delete(&perm).Error
	})
}

// UserPermissions 一次性查出该用户所有角色关联的有效接口权限。
// middleware.PermissionAuth 与子应用代理入口共用，保证口径一致。
func (s PermissionService) UserPermissions(userID uint64) ([]models.Permission, error) {
	var perms []models.Permission
	err := db.DB.
		Model(&models.Permission{}).
		Joins("JOIN base_role_permission ON base_role_permission.permission_id = base_permission.id").
		Joins("JOIN base_user_role ON base_user_role.role_id = base_role_permission.role_id").
		Where("base_user_role.user_id = ? AND base_permission.status = ?", userID, 1).
		Find(&perms).Error
	return perms, err
}

// HasAppAccess 判断用户能否访问某个子应用的接口。
//
// 口径为「按应用启用」：该 app_code 下没有登记任何接口权限点时，视为子应用不使用底座权限
// （由子应用自行鉴权），直接放行；一旦登记了权限点，就必须由角色显式授权才能访问。
func (s PermissionService) HasAppAccess(userID uint64, appCode, method, requestPath string) (bool, error) {
	var appPerms []models.Permission
	if err := db.DB.Model(&models.Permission{}).
		Where("app_code = ? AND status = ? AND method <> '' AND path <> ''", appCode, 1).
		Find(&appPerms).Error; err != nil {
		return false, err
	}
	if len(appPerms) == 0 {
		return true, nil
	}

	userPerms, err := s.UserPermissions(userID)
	if err != nil {
		return false, err
	}
	candidates := permmatch.Candidates(requestPath)
	for _, p := range userPerms {
		if p.AppCode != appCode || !strings.EqualFold(p.Method, method) {
			continue
		}
		for _, path := range candidates {
			if permmatch.Match(p.Path, path) {
				return true, nil
			}
		}
	}
	return false, nil
}

func (s PermissionService) List(appCode string) ([]models.Permission, error) {
	var list []models.Permission
	query := db.DB.Model(&models.Permission{})
	if appCode != "" {
		query = query.Where("app_code = ?", appCode)
	}
	err := query.Order("parent_id ASC, created_at ASC").Find(&list).Error
	return list, err
}

func (s PermissionService) GetTree(appCode string) ([]models.Permission, error) {
	list, err := s.List(appCode)
	if err != nil {
		return nil, err
	}
	return buildPermTree(list, 0), nil
}

func buildPermTree(list []models.Permission, parentID uint64) []models.Permission {
	var tree []models.Permission
	for _, p := range list {
		if p.ParentID == parentID {
			children := buildPermTree(list, p.ID)
			if len(children) > 0 {
				p.Children = make([]models.Permission, len(children))
				copy(p.Children, children)
			}
			tree = append(tree, p)
		}
	}
	return tree
}
