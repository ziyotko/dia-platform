package services

import (
	"server/models"
	"server/utils"
)

type CategoryService struct{}

func (s *CategoryService) GetCategories(name string, status int, page int, pageSize int) ([]models.Category, int64, error) {
	var categories []models.Category
	var total int64
	query := utils.DB.Model(&models.Category{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err = query.Order("sort ASC, id DESC").Limit(pageSize).Offset(offset).Find(&categories).Error
	return categories, total, err
}

func (s *CategoryService) GetAllCategories() ([]models.Category, error) {
	var categories []models.Category
	err := utils.DB.Where("status = ?", 1).Order("sort ASC, id DESC").Find(&categories).Error
	return categories, err
}

func (s *CategoryService) CreateCategory(category *models.Category) error {
	return utils.DB.Create(category).Error
}

func (s *CategoryService) UpdateCategory(id uint, category *models.Category) error {
	var old models.Category
	if err := utils.DB.First(&old, id).Error; err != nil {
		return err
	}
	updates := map[string]interface{}{
		"name":        category.Name,
		"code":        category.Code,
		"description": category.Description,
		"sort":        category.Sort,
		"status":      category.Status,
	}
	return utils.DB.Model(&old).Updates(updates).Error
}

func (s *CategoryService) UpdateCategoryStatus(id uint, status int) error {
	return utils.DB.Model(&models.Category{}).Where("id = ?", id).Update("status", status).Error
}

func (s *CategoryService) DeleteCategory(id uint) error {
	var category models.Category
	if err := utils.DB.First(&category, id).Error; err != nil {
		return err
	}
	return utils.DB.Unscoped().Delete(&category).Error
}

type CategoryArticleStat struct {
	CategoryID   uint   `json:"categoryId"`
	CategoryName string `json:"categoryName"`
	Count        int64  `json:"count"`
}

func (s *CategoryService) GetCategoryArticleStats() ([]CategoryArticleStat, int64, error) {
	var results []CategoryArticleStat
	var total int64

	err := utils.DB.Model(&models.Article{}).
		Select("category_id as category_id, category.name as category_name, COUNT(*) as count").
		Joins("LEFT JOIN category ON article.category_id = category.id").
		Group("category_id").
		Scan(&results).Error
	if err != nil {
		return nil, 0, err
	}

	for _, r := range results {
		total += r.Count
	}

	return results, total, nil
}
