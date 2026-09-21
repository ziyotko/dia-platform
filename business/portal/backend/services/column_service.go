package services

import (
	"errors"
	"fmt"
	"time"

	"server/models"
	"server/utils"
)

type ColumnService struct{}

// ColumnPublishItem 栏目发布（静态化）列表项
type ColumnPublishItem struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`      // 栏目名称
	Template  string `json:"template"`  // 模板名称
	RoutePath string `json:"routePath"` // 访问路径 = 栏目 route_path + 模板页 route_path
	URL       string `json:"url"`       // 可直接访问的地址（站点地址 + 访问路径），供前端「预览」使用
	CreatedAt string `json:"createdAt"` // 模板页创建时间
	UpdatedAt string `json:"updatedAt"` // 模板页修改时间
}

// GetColumnPublishes 获取栏目发布（静态化）列表：
// 查询有已发布文章的栏目，并关联栏目页模板（type='column' and status=1）
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

	// 2. 查询栏目页模板（模板名称 + route_path + 创建/修改时间）
	type ColumnPage struct {
		Name      string    `gorm:"column:name"`
		RoutePath string    `gorm:"column:route_path"`
		CreatedAt time.Time `gorm:"column:created_at"`
		UpdatedAt time.Time `gorm:"column:updated_at"`
	}
	var page ColumnPage
	result := utils.DB.Model(&models.Template{}).
		Select("name, route_path, created_at, updated_at").
		Where("type = ? AND status = ?", "column", 1).
		Order("id ASC").
		Limit(1).
		Scan(&page)
	if result.Error != nil {
		return nil, result.Error
	}
	// 无启用中的栏目页模板时 Scan 不会报错（只是 0 行、结构体为零值），
	// 必须用 RowsAffected 判定，否则会把 0001-01-01 零值时间输出到列表
	hasTemplate := result.RowsAffected > 0

	// 3. 拼接结果：访问路径 = 栏目 route_path + 模板页 route_path
	baseURL := SiteBaseURL()
	items := make([]ColumnPublishItem, 0, len(columns))
	for _, col := range columns {
		item := ColumnPublishItem{
			ID:        col.ID,
			Name:      col.Name,
			RoutePath: col.RoutePath,
		}
		if hasTemplate {
			item.Template = page.Name
			item.RoutePath = col.RoutePath + page.RoutePath
			item.CreatedAt = page.CreatedAt.Format("2006-01-02 15:04:05")
			item.UpdatedAt = page.UpdatedAt.Format("2006-01-02 15:04:05")
		}
		item.URL = BuildPageAccessURL(baseURL, item.RoutePath)
		items = append(items, item)
	}
	return items, nil
}

func (s *ColumnService) GetColumns(templateID uint, parentID uint, displayType int) ([]models.Column, error) {
	var columns []models.Column
	query := utils.DB.Model(&models.Column{})
	if templateID > 0 {
		query = query.Where("template_id = ?", templateID)
	}
	if parentID > 0 {
		query = query.Where("parent_id = ?", parentID)
	}
	if displayType > 0 {
		query = query.Where("display_type = ?", displayType)
	}
	err := query.Order("sort ASC, id ASC").Preload("Workflow").Find(&columns).Error
	return columns, err
}

// validateColumnBinding 校验栏目的模板归属与上级栏目：
//   - 模板必填且必须存在：栏目一律挂在模板下（column.template_id），写入 0 或无效 ID 的栏目
//     在「栏目管理」（按 templateId 过滤）中永远不可见，只能进库修正；
//   - 上级栏目必须存在且与当前栏目属于同一模板（否则跨模板挂载后父子不同页，同样不可见）。
func (s *ColumnService) validateColumnBinding(nodeID, templateID, parentID uint) error {
	if templateID == 0 {
		return errors.New("请先选择所属模板")
	}
	var tpl models.Template
	if err := utils.DB.First(&tpl, templateID).Error; err != nil {
		return errors.New("所选模板不存在")
	}
	if parentID == 0 {
		return nil
	}
	if nodeID > 0 && parentID == nodeID {
		return errors.New("上级栏目不能是自己")
	}
	var parent models.Column
	if err := utils.DB.First(&parent, parentID).Error; err != nil {
		return errors.New("上级栏目不存在")
	}
	if parent.TemplateID != templateID {
		return errors.New("上级栏目与所属模板不一致")
	}
	return nil
}

func (s *ColumnService) CreateColumn(column *models.Column) error {
	if err := s.validateColumnBinding(0, column.TemplateID, column.ParentID); err != nil {
		return err
	}
	return utils.DB.Create(column).Error
}

func (s *ColumnService) UpdateColumn(id uint, column *models.Column) error {
	if err := s.validateColumnBinding(id, column.TemplateID, column.ParentID); err != nil {
		return err
	}
	var old models.Column
	if err := utils.DB.First(&old, id).Error; err != nil {
		return err
	}
	updates := map[string]any{
		"name":         column.Name,
		"code":         column.Code,
		"template_id":  column.TemplateID,
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

// DeleteColumn 删除栏目：存在子栏目时拒绝删除。
// 子栏目的 parent_id 指向已删除的栏目，而前端栏目树只把 parent_id=0 当作根节点，
// 子栏目会从界面上彻底消失且无法再编辑/删除（前端文案原先承诺「一并删除」，实际并不级联）。
func (s *ColumnService) DeleteColumn(id uint) error {
	var column models.Column
	if err := utils.DB.First(&column, id).Error; err != nil {
		return err
	}
	var childCount int64
	if err := utils.DB.Model(&models.Column{}).Where("parent_id = ?", id).Count(&childCount).Error; err != nil {
		return err
	}
	if childCount > 0 {
		return fmt.Errorf("该栏目下仍有 %d 个子栏目，请先删除或调整子栏目后再删除", childCount)
	}
	return utils.DB.Delete(&column).Error
}
