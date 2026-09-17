package service

import (
	"errors"
	"member/internal/models"
	"member/pkg/db"

	"gorm.io/gorm"
)

type OrganizationService struct{}

// GetOrganizationTree returns the org tree
func (s *OrganizationService) GetOrganizationTree() ([]*models.Organization, error) {
	var orgs []models.Organization
	if err := db.DB.Order("sort ASC, id ASC").Find(&orgs).Error; err != nil {
		return nil, err
	}
	// Load levels for each org
	for i := range orgs {
		db.DB.Where("org_id = ?", orgs[i].ID).Preload("Level").Find(&orgs[i].Levels)
	}
	return buildTree(orgs, 0), nil
}

// GetOrganization returns an org by ID
func (s *OrganizationService) GetOrganization(id uint64) (*models.Organization, error) {
	var org models.Organization
	if err := db.DB.First(&org, id).Error; err != nil {
		return nil, errors.New("组织不存在")
	}
	return &org, nil
}

// CreateOrganization creates an org (admin)
func (s *OrganizationService) CreateOrganization(req CreateOrgRequest) (*models.Organization, error) {
	orgType := req.Type
	if orgType == "" {
		if req.ParentID == 0 {
			orgType = "branch"
		} else {
			orgType = "representative"
		}
	}

	// 只允许一个顶级机构（总会）
	if req.ParentID == 0 {
		var rootCount int64
		if err := db.DB.Model(&models.Organization{}).Where("parent_id = ?", 0).Count(&rootCount).Error; err != nil {
			return nil, err
		}
		if rootCount > 0 {
			return nil, errors.New("已存在上级机构，只允许一个顶级机构")
		}
	}

	// Validate 2-level max
	if req.ParentID > 0 {
		var parent models.Organization
		if err := db.DB.First(&parent, req.ParentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("上级机构不存在")
			}
			return nil, err
		}
		if parent.ParentID != 0 {
			return nil, errors.New("不能超过两级")
		}
	}

	org := models.Organization{
		Name:        req.Name,
		ParentID:    req.ParentID,
		Type:        orgType,
		Description: req.Description,
		ContactInfo: req.ContactInfo,
		Sort:        req.Sort,
	}
	if err := db.DB.Create(&org).Error; err != nil {
		return nil, err
	}
	return &org, nil
}

// UpdateOrganization updates an org (admin)
func (s *OrganizationService) UpdateOrganization(id uint64, req UpdateOrgRequest) error {
	// 只允许一个顶级机构：禁止把下级机构提升为顶级机构
	if req.ParentID != nil && *req.ParentID == 0 {
		var org models.Organization
		if err := db.DB.First(&org, id).Error; err != nil {
			return errors.New("组织不存在")
		}
		if org.ParentID != 0 {
			return errors.New("只允许一个顶级机构，无法将下级机构提升为顶级机构")
		}
	}
	// 局部更新：只写请求里显式传入的字段，避免「只改简介」把名称、联系方式、排序等一并清空。
	updates := map[string]interface{}{}
	if req.Name != nil {
		if *req.Name == "" {
			return errors.New("名称不能为空")
		}
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.ContactInfo != nil {
		updates["contact_info"] = *req.ContactInfo
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if req.ParentID != nil {
		updates["parent_id"] = *req.ParentID
	}
	if len(updates) == 0 {
		return nil
	}
	return db.DB.Model(&models.Organization{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteOrganization deletes an org (admin)
func (s *OrganizationService) DeleteOrganization(id uint64) error {
	var org models.Organization
	if err := db.DB.First(&org, id).Error; err != nil {
		return errors.New("组织不存在")
	}
	// Level-1 (parent_id = 0) cannot be deleted
	if org.ParentID == 0 {
		return errors.New("分支机构为一级组织，不可删除，可修改名称")
	}
	// Check for children (shouldn't happen since we limit to 2 levels)
	var count int64
	db.DB.Model(&models.Organization{}).Where("parent_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("该组织下有子组织，无法删除")
	}
	// 占用校验：被会员加入 / 入会申请 / 会费记录引用的组织不可删除，避免悬空 org_id
	var joinCount int64
	db.DB.Model(&models.MemberOrganization{}).Where("org_id = ?", id).Count(&joinCount)
	if joinCount > 0 {
		return errors.New("该组织已有会员加入，无法删除")
	}
	var appCount int64
	db.DB.Model(&models.Application{}).Where("org_id = ?", id).Count(&appCount)
	if appCount > 0 {
		return errors.New("该组织已有入会申请记录，无法删除")
	}
	var feeCount int64
	db.DB.Model(&models.FeeRecord{}).Where("org_id = ?", id).Count(&feeCount)
	if feeCount > 0 {
		return errors.New("该组织已有会费记录，无法删除")
	}
	// Clean up level associations
	db.DB.Where("org_id = ?", id).Delete(&models.MemberOrgLevel{})
	return db.DB.Delete(&models.Organization{}, id).Error
}

type CreateOrgRequest struct {
	Name        string `json:"name" binding:"required"`
	ParentID    uint64 `json:"parent_id"`
	Type        string `json:"type"`
	Description string `json:"description"`
	ContactInfo string `json:"contact_info"`
	Sort        int    `json:"sort"`
}

// UpdateOrgRequest 机构更新请求。字段均为指针，nil 表示「本次不修改该字段」。
// 注意：机构类型（type）与层级由创建时决定，不支持修改（仅两级结构）。
type UpdateOrgRequest struct {
	Name        *string `json:"name"`
	ParentID    *uint64 `json:"parent_id"`
	Description *string `json:"description"` // 简介（会员端「加入的组织机构」/首页展示）
	ContactInfo *string `json:"contact_info"`
	Sort        *int    `json:"sort"`
}

func buildTree(orgs []models.Organization, parentID uint64) []*models.Organization {
	var tree []*models.Organization
	for i := range orgs {
		if orgs[i].ParentID == parentID {
			node := &orgs[i]
			node.Children = buildTree(orgs, node.ID)
			tree = append(tree, node)
		}
	}
	return tree
}
