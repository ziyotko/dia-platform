package service

import (
	"errors"

	"application/internal/models"
	"application/pkg/db"
)

type CategoryService struct{}

func (s *CategoryService) Create(c *models.ProjectCategory) error {
	return db.DB.Create(c).Error
}

func (s *CategoryService) Update(id uint64, updates map[string]interface{}) error {
	clean := pickUpdates(updates, "name", "description", "sort")
	if len(clean) == 0 {
		return nil
	}
	return db.DB.Model(&models.ProjectCategory{}).Where("id = ?", id).Updates(clean).Error
}

func (s *CategoryService) Delete(id uint64) error {
	var count int64
	db.DB.Model(&models.ProjectBatch{}).Where("category_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("该类别下存在申报批次，无法删除")
	}
	return db.DB.Delete(&models.ProjectCategory{}, id).Error
}

func (s *CategoryService) List() ([]models.ProjectCategory, error) {
	var list []models.ProjectCategory
	err := db.DB.Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}
