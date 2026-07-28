package service

import (
	"errors"
	"member/internal/models"
	"member/pkg/db"
)

type OrganizationService struct{}

// GetOrganizationTree returns the org tree
func (s *OrganizationService) GetOrganizationTree() ([]*models.Organization, error) {
	var orgs []models.Organization
	if err := db.DB.Order("sort ASC").Find(&orgs).Error; err != nil {
		return nil, err
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
	org := models.Organization{
		Name:        req.Name,
		ParentID:    req.ParentID,
		Type:        req.Type,
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
	updates := map[string]interface{}{
		"name":         req.Name,
		"type":         req.Type,
		"description":  req.Description,
		"contact_info": req.ContactInfo,
		"sort":         req.Sort,
	}
	if req.ParentID != nil {
		updates["parent_id"] = *req.ParentID
	}
	return db.DB.Model(&models.Organization{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteOrganization deletes an org (admin)
func (s *OrganizationService) DeleteOrganization(id uint64) error {
	// Check for children
	var count int64
	db.DB.Model(&models.Organization{}).Where("parent_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("该组织下有子组织，无法删除")
	}
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

type UpdateOrgRequest struct {
	Name        string  `json:"name"`
	ParentID    *uint64 `json:"parent_id"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	ContactInfo string  `json:"contact_info"`
	Sort        int     `json:"sort"`
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
