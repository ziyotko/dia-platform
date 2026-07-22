package services

import (
	"server/models"
	"server/utils"
)

type StaticPageService struct{}

func (s *StaticPageService) GetStaticPages(pageType string, templateID uint) ([]models.Page, error) {
	var pages []models.Page
	query := utils.DB.Model(&models.Page{})
	if pageType != "" {
		query = query.Where("page_type = ?", pageType)
	}
	if templateID > 0 {
		query = query.Where("template_id = ?", templateID)
	}
	query = query.Where("status=?", 1)
	err := query.Order("id DESC").Find(&pages).Error
	return pages, err
}
