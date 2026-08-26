package services

import (
	"time"

	"server/models"
	"server/utils"
)

type ColumnService struct{}

// ColumnPublishItem 栏目发布（静态化）列表项
type ColumnPublishItem struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`       // 栏目名称
	Template   string `json:"template"`   // 模板名称
	RoutePath  string `json:"routePath"`  // 访问路径 = 栏目 route_path + 模板页 route_path
	CreateTime string `json:"createTime"` // 模板页创建时间
	UpdatedAt  string `json:"updatedAt"`  // 模板页修改时间
}

// GetColumnPublishes 获取栏目发布（静态化）列表：
// 查询有已发布文章的栏目，并关联栏目模板页（page_type='column' and status=1）
func (s *ColumnService) GetColumnPublishes() ([]ColumnPublishItem, error) {
	// 1. 查询已发布文章所属的栏目（栏目名称 + 栏目 route_path）
	type ColumnItem struct {
		ID        uint   `gorm:"column:id"`
		Name      string `gorm:"column:name"`
		RoutePath string `gorm:"column:route_path"`
	}
	var columns []ColumnItem
	err := utils.DB.Model(&models.Column{}).
		Select("id, name, route_path").
		Where("id IN (SELECT DISTINCT column_id FROM article_column_publish) and display_type<>7").
		Scan(&columns).Error
	if err != nil {
		return nil, err
	}

	// 2. 查询栏目模板页（模板名称 + route_path + 创建/修改时间）
	type ColumnPage struct {
		Name      string    `gorm:"column:name"`
		RoutePath string    `gorm:"column:route_path"`
		CreatedAt time.Time `gorm:"column:created_at"`
		UpdatedAt time.Time `gorm:"column:updated_at"`
	}
	var page ColumnPage
	err = utils.DB.Model(&models.Page{}).
		Select("name, route_path, created_at, updated_at").
		Where("page_type = ? AND status = ?", "column", 1).
		Order("id ASC").
		Limit(1).
		Scan(&page).Error
	if err != nil {
		return nil, err
	}

	// 3. 拼接结果：访问路径 = 栏目 route_path + 模板页 route_path
	result := make([]ColumnPublishItem, 0, len(columns))
	for _, col := range columns {
		result = append(result, ColumnPublishItem{
			ID:         col.ID,
			Name:       col.Name,
			Template:   page.Name,
			RoutePath:  col.RoutePath + page.RoutePath,
			CreateTime: page.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:  page.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return result, nil
}

func (s *ColumnService) GetColumns(pageID uint, parentID uint) ([]models.Column, error) {
	var columns []models.Column
	query := utils.DB.Model(&models.Column{})
	if pageID > 0 {
		query = query.Where("page_id = ?", pageID)
	}
	if parentID > 0 {
		query = query.Where("parent_id = ?", parentID)
	}
	err := query.Order("sort ASC, id ASC").Preload("Workflow").Find(&columns).Error
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
	updates := map[string]any{
		"name":         column.Name,
		"code":         column.Code,
		"page_id":      column.PageID,
		"parent_id":    column.ParentID,
		"route_path":   column.RoutePath,
		"description":  column.Description,
		"sort":         column.Sort,
		"status":       column.Status,
		"display_type": column.DisplayType,
		"workflow_id":  column.WorkflowID,
	}
	return utils.DB.Model(&old).Updates(updates).Error
}

func (s *ColumnService) DeleteColumn(id uint) error {
	var column models.Column
	if err := utils.DB.First(&column, id).Error; err != nil {
		return err
	}
	return utils.DB.Unscoped().Delete(&column).Error
}

func (s *ColumnService) DeleteColumnsByPageID(pageID uint) error {
	return utils.DB.Where("page_id = ?", pageID).Unscoped().Delete(&models.Column{}).Error
}
