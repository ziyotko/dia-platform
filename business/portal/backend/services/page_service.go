package services

import (
	"server/models"
	"server/utils"

	"gorm.io/gorm"
)

type PageService struct{}

func (s *PageService) GetPages(pageType string, templateID uint, status *int) ([]models.Page, error) {
	var pages []models.Page
	query := utils.DB.Model(&models.Page{})
	if pageType != "" {
		query = query.Where("page_type = ?", pageType)
	}
	if templateID > 0 {
		query = query.Where("template_id = ?", templateID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	err := query.Order("id DESC").Find(&pages).Error
	return pages, err
}

func (s *PageService) CreatePage(page *models.Page) error {
	return utils.DB.Create(page).Error
}

func (s *PageService) UpdatePage(id uint, page *models.Page) error {
	updates := map[string]any{
		"name":        page.Name,
		"code":        page.Code,
		"page_type":   page.PageType,
		"route_path":  page.RoutePath,
		"template_id": page.TemplateID,
		"template":    page.Template,
		"description": page.Description,
		"status":      page.Status,
	}
	return utils.DB.Model(&models.Page{}).Where("id = ?", id).Updates(updates).Error
}

func (s *PageService) DeletePage(id uint) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		var page models.Page
		if err := tx.First(&page, id).Error; err != nil {
			return err
		}
		if err := tx.Where("page_id = ?", page.ID).Unscoped().Delete(&models.Column{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Delete(&page).Error
	})
}
