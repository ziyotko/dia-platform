package service

import (
	"errors"
	"fmt"

	"base/internal/models"
	"base/pkg/db"

	"gorm.io/gorm"
)

type PermissionService struct{}

func (s PermissionService) Create(p *models.Permission) error {
	return db.DB.Create(p).Error
}

func (s PermissionService) Update(p *models.Permission) error {
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
