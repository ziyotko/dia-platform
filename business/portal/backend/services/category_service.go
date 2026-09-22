package services

import (
	"fmt"
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
	if err := ensureNameCodeUnique(&models.Category{}, "分类", category.Name, category.Code, 0, nil); err != nil {
		return err
	}
	return utils.DB.Create(category).Error
}

func (s *CategoryService) UpdateCategory(id uint, category *models.Category) error {
	var old models.Category
	if err := utils.DB.First(&old, id).Error; err != nil {
		return err
	}
	if err := ensureNameCodeUnique(&models.Category{}, "分类", category.Name, category.Code, id, nil); err != nil {
		return err
	}
	updates := map[string]any{
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
	// 仍被文章引用时拒绝删除：article_category 有外键，直接删除会返回 1451，
	// 错误被 SanitizeError 归类后只显示「删除分类失败」，管理员无从判断原因。
	var articleCount int64
	if err := utils.DB.Model(&models.ArticleCategory{}).Where("category_id = ?", id).Count(&articleCount).Error; err != nil {
		return err
	}
	if articleCount > 0 {
		return fmt.Errorf("该分类仍被 %d 篇文章使用，请先调整这些文章的分类后再删除", articleCount)
	}
	return utils.DB.Delete(&category).Error
}

type CategoryArticleStat struct {
	CategoryID   uint   `json:"categoryId"`
	CategoryName string `json:"categoryName"`
	Count        int64  `json:"count"`
}

func (s *CategoryService) GetCategoryArticleStats() ([]CategoryArticleStat, int64, error) {
	var results []CategoryArticleStat

	// 口径与标签统计（tag_service.GetTagArticleStats）保持一致：
	//   - LEFT JOIN：未被引用/无已发布文章的分类仍以 0 出现，不从统计中消失；
	//   - 只统计已发布（status=1）且未删除的文章；
	//   - total 为统计条目数（分类个数），与 /tags/stats 一致。
	err := utils.DB.Model(&models.Category{}).
		Select("category.id as category_id, category.name as category_name, COUNT(article.id) as count").
		Joins("LEFT JOIN article_category ON article_category.category_id = category.id").
		Joins("LEFT JOIN article ON article.id = article_category.article_id AND article.status = ?", models.ArticleStatusPublished).
		Group("category.id, category.name").
		Order("count DESC, category.id ASC").
		Scan(&results).Error
	if err != nil {
		return nil, 0, err
	}

	return results, int64(len(results)), nil
}
