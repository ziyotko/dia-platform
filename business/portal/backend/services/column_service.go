package services

import (
	"errors"
	"fmt"
	"time"

	"server/models"
	"server/utils"

	"gorm.io/gorm"
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
			item.RoutePath = JoinAccessPath(col.RoutePath, page.RoutePath)
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
	// 上级栏目不能是自身的下级（否则父子成环，整枝栏目都从「栏目管理」/投放树中消失，只能进库修正）。
	// 前端父级下拉展示的是整个模板树（含自身与后代），故必须在后端拦下。
	// 注：此处不能直接复用 tree_guard.validateTreeParent——栏目表名为 MySQL 保留字 `column`，
	// 且它是按物理表名拼 SQL 的，因此改为按模型向父级追溯。
	if nodeID > 0 {
		if err := ensureColumnNotDescendant(nodeID, parentID); err != nil {
			return err
		}
	}
	return nil
}

// ensureColumnNotDescendant 确认 nodeID 不是 parentID 的祖先（即 parentID 不是 nodeID 的下级）。
func ensureColumnNotDescendant(nodeID, parentID uint) error {
	cur := parentID
	seen := map[uint]bool{parentID: true}
	for cur != 0 {
		if cur == nodeID {
			return errors.New("上级栏目不能是自己或其下级栏目")
		}
		var node models.Column
		if err := utils.DB.Select("id, parent_id").First(&node, cur).Error; err != nil {
			// 只有「记录确实不存在」才视为数据异常不阻断；其它错误（DB 抖动/超时）必须上抛，
			// 否则会把「查不到祖先」当成「不是后代」而放行成环（成环后整枝栏目从界面消失）。
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if node.ParentID == 0 || seen[node.ParentID] {
			return nil
		}
		seen[node.ParentID] = true
		cur = node.ParentID
	}
	return nil
}

func (s *ColumnService) CreateColumn(column *models.Column) error {
	if err := s.validateColumnBinding(0, column.TemplateID, column.ParentID); err != nil {
		return err
	}
	return utils.DB.Create(column).Error
}

// columnWorkflowID 取栏目绑定的审核流程 ID（nil 与 0 等价，均表示未绑定流程）
func columnWorkflowID(id *uint) uint {
	if id == nil {
		return 0
	}
	return *id
}

func (s *ColumnService) UpdateColumn(id uint, column *models.Column) error {
	if err := s.validateColumnBinding(id, column.TemplateID, column.ParentID); err != nil {
		return err
	}
	var old models.Column
	if err := utils.DB.First(&old, id).Error; err != nil {
		return err
	}
	// 更换所属模板时必须先清空子栏目：子栏目的 parent_id 仍指向本栏目而 template_id 还是旧模板，
	// 前端栏目树按「同模板 + parent_id」过滤 → 子栏目既不是根节点、父节点又不在当前模板的列表里，
	// 会从「栏目管理」中彻底消失（只能进库修正）。
	if old.TemplateID != column.TemplateID {
		var childCount int64
		if err := utils.DB.Model(&models.Column{}).Where("parent_id = ?", id).Count(&childCount).Error; err != nil {
			return err
		}
		if childCount > 0 {
			return fmt.Errorf("该栏目下仍有 %d 个子栏目，不能更换所属模板（更换后子栏目会因父子模板不一致而不可见），请先删除或迁移子栏目", childCount)
		}
	}
	// 审核中的文章按 article_column_audit.workflow_id 推进，改栏目流程会让界面展示的流程与实际执行的不一致。
	if columnWorkflowID(old.WorkflowID) != columnWorkflowID(column.WorkflowID) {
		var pending int64
		if err := utils.DB.Model(&models.ArticleColumnAudit{}).
			Where("column_id = ? AND status = ?", id, 0).
			Count(&pending).Error; err != nil {
			return err
		}
		if pending > 0 {
			return fmt.Errorf("该栏目下有 %d 篇文章正在审核，不能修改审核流程，请等待审核结束后再改", pending)
		}
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
	// 广告/友链通过 template_id + column_id 关联栏目（无外键约束）：直接删除会留下悬挂 column_id，
	// 「位置」列显示为空，且这些记录下次编辑保存时会被 ValidateTemplateColumn 判为「所选栏目不属于该模板」
	// 而无法保存（与 DeleteTemplate 的处置口径一致）。
	var adCount, linkCount int64
	if err := utils.DB.Model(&models.Ad{}).Where("column_id = ?", id).Count(&adCount).Error; err != nil {
		return err
	}
	if err := utils.DB.Model(&models.Link{}).Where("column_id = ?", id).Count(&linkCount).Error; err != nil {
		return err
	}
	if adCount > 0 || linkCount > 0 {
		return fmt.Errorf("该栏目仍被 %d 条广告、%d 条友链使用，请先调整这些记录后再删除栏目", adCount, linkCount)
	}
	return utils.DB.Delete(&column).Error
}
