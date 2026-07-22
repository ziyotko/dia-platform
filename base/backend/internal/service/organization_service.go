package service

import (
	"base/internal/models"
	"base/pkg/db"
)

type OrganizationService struct{}

func (s OrganizationService) Create(o *models.Organization) error {
	return db.DB.Create(o).Error
}

func (s OrganizationService) Update(o *models.Organization, tenantID uint64) error {
	db := db.DB.Model(o)
	if tenantID > 0 {
		db = db.Where("tenant_id = ?", tenantID)
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

func (s OrganizationService) Delete(id uint64, tenantID uint64) error {
	db := db.DB
	if tenantID > 0 {
		db = db.Where("tenant_id = ?", tenantID)
	}
	return db.Delete(&models.Organization{BaseModel: models.BaseModel{ID: id}}).Error
}

func (s OrganizationService) GetByID(id uint64, tenantID uint64) (*models.Organization, error) {
	var o models.Organization
	query := db.DB.Where("id = ?", id)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	err := query.First(&o).Error
	return &o, err
}

func (s OrganizationService) List(tenantID uint64) ([]models.Organization, error) {
	var list []models.Organization
	err := db.DB.Model(&models.Organization{}).Where("tenant_id = ?", tenantID).Order("sort ASC, created_at ASC").Find(&list).Error
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
