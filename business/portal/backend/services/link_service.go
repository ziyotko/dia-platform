package services

import (
	"server/models"
	"server/utils"
)

type LinkService struct{}

func (s *LinkService) GetLinks(name string, pageID int, columnID int, status int, page int, pageSize int) ([]models.Link, int64, error) {
	var links []models.Link
	var total int64
	query := utils.DB.Model(&models.Link{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if pageID > 0 {
		query = query.Where("page_id = ?", pageID)
	}
	if columnID > 0 {
		query = query.Where("column_id = ?", columnID)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err = query.Order("sort ASC, id DESC").Limit(pageSize).Offset(offset).Find(&links).Error
	return links, total, err
}

func (s *LinkService) GetLinkByID(id uint) (*models.Link, error) {
	var link models.Link
	err := utils.DB.First(&link, id).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (s *LinkService) CreateLink(link *models.Link) error {
	return utils.DB.Create(link).Error
}

func (s *LinkService) UpdateLink(id uint, link *models.Link) error {
	var old models.Link
	if err := utils.DB.First(&old, id).Error; err != nil {
		return err
	}
	updates := map[string]any{
		"name":        link.Name,
		"url":         link.Url,
		"logo":        link.Logo,
		"description": link.Description,
		"page_id":     link.PageID,
		"column_id":   link.ColumnID,
		"sort":        link.Sort,
		"status":      link.Status,
	}
	return utils.DB.Model(&old).Updates(updates).Error
}

func (s *LinkService) UpdateLinkStatus(id uint, status int) error {
	return utils.DB.Model(&models.Link{}).Where("id = ?", id).Update("status", status).Error
}

func (s *LinkService) DeleteLink(id uint) error {
	var link models.Link
	if err := utils.DB.First(&link, id).Error; err != nil {
		return err
	}
	return utils.DB.Unscoped().Delete(&link).Error
}
