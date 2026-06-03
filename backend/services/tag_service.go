package services

import (
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
	updates := map[string]interface{}{
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
	return utils.DB.Delete(&tag).Error
}
