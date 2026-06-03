package services

import (
	"server/models"
	"server/utils"
)

type ColumnService struct{}

func (s *ColumnService) GetColumns(pageID uint, parentID uint) ([]models.Column, error) {
	var columns []models.Column
	query := utils.DB.Model(&models.Column{})
	if pageID > 0 {
		query = query.Where("page_id = ?", pageID)
	}
	if parentID > 0 {
		query = query.Where("parent_id = ?", parentID)
	}
	err := query.Order("sort ASC, id ASC").Find(&columns).Error
	return columns, err
}

func (s *ColumnService) CreateColumn(column *models.Column) error {
	return utils.DB.Create(column).Error
}

func (s *ColumnService) UpdateColumn(id uint, column *models.Column) error {
	var old models.Column
	if err := utils.DB.First(&old, id).Error; err != nil {
		return err
	}
	updates := map[string]interface{}{
		"name":        column.Name,
		"code":        column.Code,
		"page_id":     column.PageID,
		"parent_id":   column.ParentID,
		"route_path":  column.RoutePath,
		"template":    column.Template,
		"description": column.Description,
		"sort":        column.Sort,
		"status":      column.Status,
	}
	return utils.DB.Model(&old).Updates(updates).Error
}

func (s *ColumnService) DeleteColumn(id uint) error {
	var column models.Column
	if err := utils.DB.First(&column, id).Error; err != nil {
		return err
	}
	return utils.DB.Delete(&column).Error
}

func (s *ColumnService) DeleteColumnsByPageID(pageID uint) error {
	return utils.DB.Where("page_id = ?", pageID).Delete(&models.Column{}).Error
}
