package service

import (
	"base/internal/models"
	"base/pkg/db"
)

type PermissionService struct{}

func (s PermissionService) Create(p *models.Permission) error {
	return db.DB.Create(p).Error
}

func (s PermissionService) Update(p *models.Permission) error {
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

func (s PermissionService) Delete(id uint64) error {
	return db.DB.Delete(&models.Permission{BaseModel: models.BaseModel{ID: id}}).Error
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
