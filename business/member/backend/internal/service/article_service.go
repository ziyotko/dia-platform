package service

import (
	"errors"
	"member/internal/models"
	"member/pkg/db"
	"time"
)

type ArticleService struct{}

// CreateArticle creates a new article
func (s *ArticleService) CreateArticle(memberID uint64, req CreateArticleRequest) (*models.Article, error) {
	article := models.Article{
		MemberID:   memberID,
		CategoryID: req.CategoryID,
		Title:      req.Title,
		Content:    req.Content,
		Summary:    req.Summary,
		CoverImage: req.CoverImage,
		Status:     models.ArticleStatusDraft,
	}
	if req.Submit {
		article.Status = models.ArticleStatusPending
	}
	if err := db.DB.Create(&article).Error; err != nil {
		return nil, err
	}
	return &article, nil
}

// UpdateArticle updates an article
func (s *ArticleService) UpdateArticle(memberID, articleID uint64, req UpdateArticleRequest) error {
	var article models.Article
	if err := db.DB.Where("id = ? AND member_id = ?", articleID, memberID).First(&article).Error; err != nil {
		return errors.New("文章不存在")
	}

	updates := map[string]interface{}{
		"title":       req.Title,
		"content":     req.Content,
		"summary":     req.Summary,
		"cover_image": req.CoverImage,
		"category_id": req.CategoryID,
	}
	if req.Submit && article.Status == models.ArticleStatusDraft {
		updates["status"] = models.ArticleStatusPending
	}
	return db.DB.Model(&article).Updates(updates).Error
}

// DeleteArticle deletes an article
func (s *ArticleService) DeleteArticle(memberID, articleID uint64) error {
	return db.DB.Where("id = ? AND member_id = ?", articleID, memberID).Delete(&models.Article{}).Error
}

// AdminDeleteArticle deletes any article (admin)
func (s *ArticleService) AdminDeleteArticle(id uint64) error {
	return db.DB.Delete(&models.Article{}, id).Error
}

// GetArticle returns an article by ID (owner only)
func (s *ArticleService) GetArticle(id, memberID uint64) (*models.Article, error) {
	var article models.Article
	if err := db.DB.Preload("Member").Preload("Category").Where("id = ? AND member_id = ?", id, memberID).First(&article).Error; err != nil {
		return nil, errors.New("文章不存在")
	}
	return &article, nil
}

// GetMyArticles returns member's articles with filters
func (s *ArticleService) GetMyArticles(memberID uint64, page, size int, status string, categoryID uint64) ([]models.Article, int64, error) {
	var articles []models.Article
	var total int64

	query := db.DB.Model(&models.Article{}).Preload("Category").Where("member_id = ?", memberID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&articles).Error; err != nil {
		return nil, 0, err
	}
	return articles, total, nil
}

// GetPublishedArticle returns a single published article (public)
func (s *ArticleService) GetPublishedArticle(id uint64) (*models.Article, error) {
	var article models.Article
	if err := db.DB.Preload("Member").Preload("Category").
		Where("id = ? AND status = ?", id, models.ArticleStatusPublished).First(&article).Error; err != nil {
		return nil, errors.New("文章不存在")
	}
	return &article, nil
}

// ListArticles lists published articles (public)
func (s *ArticleService) ListArticles(page, size int, categoryID uint64, keyword string) ([]models.Article, int64, error) {
	var articles []models.Article
	var total int64

	query := db.DB.Model(&models.Article{}).Preload("Member").Preload("Category").
		Where("status = ?", models.ArticleStatusPublished)
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	if keyword != "" {
		query = query.Where("title LIKE ? OR summary LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("published_at DESC").Offset((page - 1) * size).Limit(size).Find(&articles).Error; err != nil {
		return nil, 0, err
	}
	return articles, total, nil
}

// ReviewArticle reviews a pending article (admin)
func (s *ArticleService) ReviewArticle(id uint64, approved bool, comment string) error {
	var article models.Article
	if err := db.DB.First(&article, id).Error; err != nil {
		return errors.New("文章不存在")
	}
	if article.Status != models.ArticleStatusPending {
		return errors.New("该文章不在待审核状态")
	}

	newStatus := models.ArticleStatusRejected
	updates := map[string]interface{}{
		"status":         newStatus,
		"review_comment": comment,
	}
	if approved {
		newStatus = models.ArticleStatusPublished
		updates["status"] = newStatus
		now := time.Now()
		updates["published_at"] = &models.LocalTime{Time: now}
	}
	return db.DB.Model(&article).Updates(updates).Error
}

// ListAllArticles lists all articles (admin)
func (s *ArticleService) ListAllArticles(page, size int, status string) ([]models.Article, int64, error) {
	var articles []models.Article
	var total int64

	query := db.DB.Model(&models.Article{}).Preload("Member").Preload("Category")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&articles).Error; err != nil {
		return nil, 0, err
	}
	return articles, total, nil
}

type CreateArticleRequest struct {
	CategoryID uint64 `json:"category_id"`
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content"`
	Summary    string `json:"summary"`
	CoverImage string `json:"cover_image"`
	Submit     bool   `json:"submit"`
}

type UpdateArticleRequest struct {
	CategoryID uint64 `json:"category_id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Summary    string `json:"summary"`
	CoverImage string `json:"cover_image"`
	Submit     bool   `json:"submit"`
}
