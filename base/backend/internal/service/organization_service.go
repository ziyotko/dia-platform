package service

import (
	"errors"
	"fmt"

	"base/internal/models"
	"base/pkg/db"

	"gorm.io/gorm"
)

type OrganizationService struct{}

func (s OrganizationService) Create(o *models.Organization) error {
	return db.DB.Create(o).Error
}

func (s OrganizationService) Update(o *models.Organization, tenantID uint64) error {
	check := db.DB.Model(&models.Organization{}).Where("id = ?", o.ID)
	db := db.DB.Model(o)
	if tenantID > 0 {
		check = check.Where("tenant_id = ?", tenantID)
		db = db.Where("tenant_id = ?", tenantID)
	}
	if err := ensureRecordExists(check, msgNotOwnedOrMissing); err != nil {
		return err
	}
	return db.Updates(map[string]interface{}{
		"parent_id":   o.ParentID,
		"code":        o.Code,
		"name":        o.Name,
		"leader":      o.Leader,
		"phone":       o.Phone,
		"email":       o.Email,
		"sort":        o.Sort,
		"status":      o.Status,
		"description": o.Description,
	}).Error
}

// Delete 删除机构：存在子机构或已被用户引用时拒绝删除，避免机构树出现孤儿节点。
func (s OrganizationService) Delete(id uint64, tenantID uint64) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		query := tx.Where("id = ?", id)
		if tenantID > 0 {
			query = query.Where("tenant_id = ?", tenantID)
		}
		var org models.Organization
		if err := query.First(&org).Error; err != nil {
			return errors.New(msgNotOwnedOrMissingDelete)
		}
		var children int64
		if err := tx.Model(&models.Organization{}).Where("parent_id = ?", org.ID).Count(&children).Error; err != nil {
			return err
		}
		if children > 0 {
			return fmt.Errorf("该机构下还有 %d 个子机构，请先删除子机构", children)
		}
		var members int64
		if err := tx.Model(&models.User{}).Where("organization_id = ?", org.ID).Count(&members).Error; err != nil {
			return err
		}
		if members > 0 {
			return fmt.Errorf("该机构下还有 %d 个用户，请先在「用户管理」中调整其所属机构", members)
		}
		return tx.Delete(&org).Error
	})
}

func (s OrganizationService) GetByID(id uint64, tenantID uint64) (*models.Organization, error) {
	var o models.Organization
	query := db.DB.Where("id = ?", id)
	// 读取口径：本租户 + 平台内置（tenant_id = 0）均可读；写操作仍限本租户
	if tenantID > 0 {
		query = query.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}
	err := query.First(&o).Error
	return &o, err
}

func (s OrganizationService) List(tenantID uint64) ([]models.Organization, error) {
	var list []models.Organization
	query := db.DB.Model(&models.Organization{})
	// 读取口径：本租户 + 平台内置（tenant_id = 0），与菜单/消息保持一致
	if tenantID > 0 {
		query = query.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}
	err := query.Order("sort ASC, created_at ASC").Find(&list).Error
	return list, err
}

func (s OrganizationService) GetTree(tenantID uint64) ([]models.Organization, error) {
	list, err := s.List(tenantID)
	if err != nil {
		return nil, err
	}
	return buildOrgTree(list, 0), nil
}

func buildOrgTree(list []models.Organization, parentID uint64) []models.Organization {
	var tree []models.Organization
	for _, o := range list {
		if o.ParentID == parentID {
			children := buildOrgTree(list, o.ID)
			if len(children) > 0 {
				o.Children = children
			}
			tree = append(tree, o)
		}
	}
	return tree
}
