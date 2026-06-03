package services

import (
	"server/models"
	"server/utils"

	"gorm.io/gorm"
)

type PageService struct{}

func (s *PageService) GetPages(pageType string, templateID uint) ([]models.Page, error) {
	var pages []models.Page
	query := utils.DB.Model(&models.Page{})
	if pageType != "" {
		query = query.Where("page_type = ?", pageType)
	}
	if templateID > 0 {
		query = query.Where("template_id = ?", templateID)
	}
	err := query.Order("id DESC").Find(&pages).Error
	return pages, err
}

func (s *PageService) CreatePage(page *models.Page) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(page).Error; err != nil {
			return err
		}
		if page.TemplateID > 0 {
			return tx.Model(&models.Template{}).Where("id = ?", page.TemplateID).Update("page_count", gorm.Expr("page_count + 1")).Error
		}
		return nil
	})
}

func (s *PageService) UpdatePage(id uint, page *models.Page) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		var old models.Page
		if err := tx.First(&old, id).Error; err != nil {
			return err
		}
		updates := map[string]interface{}{
			"name":        page.Name,
			"code":        page.Code,
			"page_type":   page.PageType,
			"route_path":  page.RoutePath,
			"template_id": page.TemplateID,
			"template":    page.Template,
			"description": page.Description,
			"status":      page.Status,
		}
		if err := tx.Model(&old).Updates(updates).Error; err != nil {
			return err
		}
		if old.TemplateID != page.TemplateID {
			if old.TemplateID > 0 {
				tx.Model(&models.Template{}).Where("id = ?", old.TemplateID).Update("page_count", gorm.Expr("page_count - 1"))
			}
			if page.TemplateID > 0 {
				tx.Model(&models.Template{}).Where("id = ?", page.TemplateID).Update("page_count", gorm.Expr("page_count + 1"))
			}
		}
		return nil
	})
}

func (s *PageService) DeletePage(id uint) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		var page models.Page
		if err := tx.First(&page, id).Error; err != nil {
			return err
		}
		if err := tx.Where("page_id = ?", page.ID).Delete(&models.Column{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&page).Error; err != nil {
			return err
		}
		if page.TemplateID > 0 {
			return tx.Model(&models.Template{}).Where("id = ?", page.TemplateID).Update("page_count", gorm.Expr("page_count - 1")).Error
		}
		return nil
	})
}
