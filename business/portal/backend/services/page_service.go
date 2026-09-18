package services

import (
	"errors"
	"fmt"

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

// ensureTemplateAvailable 校验「一个模板只能应用一个页面」：
//   - templateID 为 0 表示不绑定模板，直接放行；
//   - 模板必须存在（未删除）；
//   - 该模板不得已被其他页面（id != pageID）绑定；pageID 传 0 表示新增页面。
func (s *PageService) ensureTemplateAvailable(pageID, templateID uint) error {
	if templateID == 0 {
		return nil
	}

	var tpl models.Template
	if err := utils.DB.First(&tpl, templateID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("所选模板不存在，请重新选择")
		}
		return err
	}

	var usedBy models.Page
	query := utils.DB.Where("template_id = ?", templateID)
	if pageID > 0 {
		query = query.Where("id <> ?", pageID)
	}
	switch err := query.First(&usedBy).Error; {
	case err == nil:
		return fmt.Errorf("模板「%s」已被页面「%s」应用，一个模板只能应用一个页面", tpl.Name, usedBy.Name)
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil
	default:
		return err
	}
}

// ensureTemplateBindingChange 仅在本次请求要改变页面的模板绑定关系时才校验，
// 避免历史数据中已存在的重复绑定导致「改名称/切状态」也被拒绝。
func (s *PageService) ensureTemplateBindingChange(id, templateID uint) error {
	var current models.Page
	if err := utils.DB.First(&current, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("页面不存在")
		}
		return err
	}
	if current.TemplateID == templateID {
		return nil
	}
	return s.ensureTemplateAvailable(id, templateID)
}

func (s *PageService) CreatePage(page *models.Page) error {
	if err := s.ensureTemplateAvailable(0, page.TemplateID); err != nil {
		return err
	}
	return utils.DB.Create(page).Error
}

func (s *PageService) UpdatePage(id uint, page *models.Page) error {
	if err := s.ensureTemplateBindingChange(id, page.TemplateID); err != nil {
		return err
	}
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
