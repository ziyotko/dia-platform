package services

import (
	"fmt"
	"server/models"
	"server/utils"
)

type TagService struct{}

func (s *TagService) GetTags(name string, status int, page int, pageSize int) ([]models.Tag, int64, error) {
	var tags []models.Tag
	var total int64
	query := utils.DB.Model(&models.Tag{})
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
	err = query.Order("id DESC").Limit(pageSize).Offset(offset).Find(&tags).Error
	return tags, total, err
}

func (s *TagService) GetAllTags() ([]models.Tag, error) {
	var tags []models.Tag
	err := utils.DB.Where("status = ?", 1).Order("id DESC").Find(&tags).Error
	return tags, err
}

func (s *TagService) CreateTag(tag *models.Tag) error {
	return utils.DB.Create(tag).Error
}

func (s *TagService) UpdateTag(id uint, tag *models.Tag) error {
	var old models.Tag
	if err := utils.DB.First(&old, id).Error; err != nil {
		return err
	}
	updates := map[string]any{
		"name":   tag.Name,
		"color":  tag.Color,
		"status": tag.Status,
	}
	return utils.DB.Model(&old).Updates(updates).Error
}

func (s *TagService) UpdateTagStatus(id uint, status int) error {
	return utils.DB.Model(&models.Tag{}).Where("id = ?", id).Update("status", status).Error
}

func (s *TagService) DeleteTag(id uint) error {
	var tag models.Tag
	if err := utils.DB.First(&tag, id).Error; err != nil {
		return err
	}
	// 仍被文章引用时拒绝删除：article_tag 是 many2many 关联表（无模型），
	// 直接删除会命中外键 1451，只能得到笼统的「删除标签失败」。
	var articleCount int64
	if err := utils.DB.Table("article_tag").Where("tag_id = ?", id).Count(&articleCount).Error; err != nil {
		return err
	}
	if articleCount > 0 {
		return fmt.Errorf("该标签仍被 %d 篇文章使用，请先调整这些文章的标签后再删除", articleCount)
	}
	return utils.DB.Delete(&tag).Error
}

type TagArticleStat struct {
	TagID uint   `json:"tagId"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Count int64  `json:"count"`
}

func (s *TagService) GetTagArticleStats() ([]TagArticleStat, error) {
	var results []TagArticleStat
	// 统计口径与文章列表/作者统计一致：只统计已发布（status=1）且未删除的文章；
	// 保留 LEFT JOIN，使未被引用的标签仍以 0 出现在标签云中。
	err := utils.DB.Model(&models.Tag{}).
		Select("tag.id as tag_id, tag.name, tag.color, COUNT(article.id) as count").
		Joins("LEFT JOIN article_tag ON article_tag.tag_id = tag.id").
		Joins("LEFT JOIN article ON article.id = article_tag.article_id AND article.status = ?", models.ArticleStatusPublished).
		Group("tag.id, tag.name, tag.color").
		Order("count DESC, tag.id ASC").
		Scan(&results).Error
	return results, err
}
