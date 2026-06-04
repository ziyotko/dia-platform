package services

import (
	"server/models"
	"server/utils"

	"gorm.io/gorm"
)

type ArticleService struct{}

func (s *ArticleService) GetArticles(title string, categoryID int, status int, page int, pageSize int) ([]models.Article, int64, error) {
	var articles []models.Article
	var total int64
	query := utils.DB.Model(&models.Article{})
	if title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err = query.Preload("Category").Preload("Tags").Order("is_top DESC, id DESC").Limit(pageSize).Offset(offset).Find(&articles).Error
	return articles, total, err
}

func (s *ArticleService) GetArticleByID(id uint) (*models.Article, error) {
	var article models.Article
	err := utils.DB.Preload("Category").Preload("Tags").First(&article, id).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (s *ArticleService) CreateArticle(article *models.Article, tagIDs []uint) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(article).Error; err != nil {
			return err
		}
		if len(tagIDs) > 0 {
			var tags []models.Tag
			for _, id := range tagIDs {
				tags = append(tags, models.Tag{ID: id})
			}
			if err := tx.Model(article).Association("Tags").Append(&tags); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *ArticleService) UpdateArticle(id uint, article *models.Article, tagIDs []uint) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		var old models.Article
		if err := tx.First(&old, id).Error; err != nil {
			return err
		}
		updates := map[string]interface{}{
			"title":       article.Title,
			"category_id": article.CategoryID,
			"summary":     article.Summary,
			"content":     article.Content,
			"status":      article.Status,
			"is_top":      article.IsTop,
			"cover":       article.Cover,
			"author":      article.Author,
			"author_code": article.AuthorCode,
			"source":      article.Source,
		}
		if err := tx.Model(&old).Updates(updates).Error; err != nil {
			return err
		}
		if len(tagIDs) > 0 {
			var tags []models.Tag
			for _, tid := range tagIDs {
				tags = append(tags, models.Tag{ID: tid})
			}
			if err := tx.Model(&old).Association("Tags").Replace(&tags); err != nil {
				return err
			}
		} else {
			if err := tx.Model(&old).Association("Tags").Clear(); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *ArticleService) UpdateArticleStatus(id uint, status int) error {
	return utils.DB.Model(&models.Article{}).Where("id = ?", id).Update("status", status).Error
}

func (s *ArticleService) DeleteArticle(id uint) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		var article models.Article
		if err := tx.First(&article, id).Error; err != nil {
			return err
		}
		if err := tx.Model(&article).Association("Tags").Clear(); err != nil {
			return err
		}
		return tx.Delete(&article).Error
	})
}

func (s *ArticleService) GetArticleCountByAuthor(author string) int64 {
	var count int64
	utils.DB.Model(&models.Article{}).Where("author = ?", author).Count(&count)
	return count
}
