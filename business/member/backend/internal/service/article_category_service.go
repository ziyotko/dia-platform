package service

import (
	"errors"
	"member/internal/models"
	"member/pkg/db"
)

type ArticleCategoryService struct{}

// ListCategories returns all article categories
func (s *ArticleCategoryService) ListCategories() ([]models.ArticleCategory, error) {
	var categories []models.ArticleCategory
	if err := db.DB.Order("sort ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// CreateCategory creates a category (admin)
func (s *ArticleCategoryService) CreateCategory(name string, sort int) (*models.ArticleCategory, error) {
	cat := models.ArticleCategory{
		Name: name,
		Sort: sort,
	}
	if err := db.DB.Create(&cat).Error; err != nil {
		return nil, err
	}
	return &cat, nil
}

// UpdateCategory updates a category (admin)
func (s *ArticleCategoryService) UpdateCategory(id uint64, name string, sort int) error {
	return db.DB.Model(&models.ArticleCategory{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name": name,
		"sort": sort,
	}).Error
}

// DeleteCategory deletes a category (admin)
func (s *ArticleCategoryService) DeleteCategory(id uint64) error {
	var cat models.ArticleCategory
	if err := db.DB.First(&cat, id).Error; err != nil {
		return errors.New("分类不存在")
	}
	// Check if articles use this category
	var count int64
	if err := db.DB.Model(&models.Article{}).Where("category_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该分类下有文章，无法删除")
	}
	return db.DB.Delete(&cat).Error
}
