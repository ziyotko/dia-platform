package services

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"server/models"
	"server/utils"
)

// 模板类型（与前端 首页/栏目页/详情页/专题页 一致）。
const (
	templateTypeColumn = "column"
	templateTypeDetail = "detail"
)

type TemplateService struct{}

type TemplateListResult struct {
	Total        int64             `json:"total"`
	List         []models.Template `json:"list"`
	ColumnCounts map[uint]int      `json:"-"`
	ColumnNames  map[uint]string   `json:"-"`
}

func (s *TemplateService) GetTemplateList(page, pageSize int, name, ttype string) (*TemplateListResult, error) {
	var list []models.Template
	var total int64

	query := utils.DB.Model(&models.Template{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if ttype != "" {
		query = query.Where("type = ?", ttype)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	// pageSize <= 0 表示不限制（供下拉选项使用，避免被分页截断）
	listQuery := query.Order("type,id DESC")
	if pageSize > 0 {
		listQuery = listQuery.Offset((page - 1) * pageSize).Limit(pageSize)
	}
	err = listQuery.Find(&list).Error
	if err != nil {
		return nil, err
	}

	// 实时统计每个模板下的栏目：供删除前的占用提示使用（页面层已合并进模板，模板直接承载栏目）
	columnCounts := make(map[uint]int)
	columnNames := make(map[uint]string)
	if len(list) > 0 {
		ids := make([]uint, len(list))
		for i, t := range list {
			ids[i] = t.ID
		}

		var columns []models.Column
		if err := utils.DB.Model(&models.Column{}).
			Select("template_id, name").
			Where("template_id IN ?", ids).
			Order("id").
			Find(&columns).Error; err == nil {
			for _, c := range columns {
				columnCounts[c.TemplateID]++
				if columnNames[c.TemplateID] == "" {
					columnNames[c.TemplateID] = c.Name
				}
			}
		}
	}

	return &TemplateListResult{
		Total:        total,
		List:         list,
		ColumnCounts: columnCounts,
		ColumnNames:  columnNames,
	}, nil
}

// isSingleActiveTemplateType 判断该模板类型是否受「同时只能启用一个」约束。
// 与页面口径一致：仅栏目页 / 详情页受限，首页 / 专题页不受限。
func isSingleActiveTemplateType(templateType string) bool {
	return templateType == templateTypeColumn || templateType == templateTypeDetail
}

// disableOtherActiveTemplates 禁用同类型的其他启用模板（仅栏目页 / 详情页生效）。
func disableOtherActiveTemplates(tx *gorm.DB, templateType string, keepID uint) error {
	if !isSingleActiveTemplateType(templateType) {
		return nil
	}
	return tx.Model(&models.Template{}).
		Where("type = ? AND id <> ? AND status = ?", templateType, keepID, 1).
		Update("status", 0).Error
}

// enforceSingleActiveTemplate 对已落库的模板应用「栏目页/详情页模板同时只能启用一个」：
// 仅当该模板当前为启用状态时，才禁用同类型的其他启用模板。
func enforceSingleActiveTemplate(tx *gorm.DB, id uint) error {
	var tpl models.Template
	if err := tx.First(&tpl, id).Error; err != nil {
		return nil // 记录不存在（如已被删除）时不处理
	}
	if tpl.Status != 1 {
		return nil
	}
	return disableOtherActiveTemplates(tx, tpl.Type, tpl.ID)
}

func (s *TemplateService) CreateTemplate(template *models.Template) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(template).Error; err != nil {
			return err
		}
		if template.Status != 1 {
			return nil
		}
		return disableOtherActiveTemplates(tx, template.Type, template.ID)
	})
}

// UpdateTemplate 更新模板。写入后统一应用「栏目页/详情页模板同时只能启用一个」约束，
// 避免从编辑信息 / 保存设计等入口绕过状态开关，造成同类型出现多个启用模板。
func (s *TemplateService) UpdateTemplate(id uint, updates map[string]any) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Template{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		return enforceSingleActiveTemplate(tx, id)
	})
}

// UpdateTemplateStatus 更新模板启用状态：栏目页 / 详情页模板同时只能启用一个，
// 启用某个模板时自动禁用同类型的其他启用模板（口径与页面一致）。
func (s *TemplateService) UpdateTemplateStatus(id uint, status int) error {
	var tpl models.Template
	if err := utils.DB.First(&tpl, id).Error; err != nil {
		return errors.New("模板不存在")
	}
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Template{}).Where("id = ?", tpl.ID).Update("status", status).Error; err != nil {
			return err
		}
		if status != 1 {
			return nil
		}
		return disableOtherActiveTemplates(tx, tpl.Type, tpl.ID)
	})
}

// DeleteTemplate 删除模板（物理删除）。
// 模板直接承载栏目（column.template_id），仍有栏目时拒绝删除，
// 否则栏目会因失去归属而在「栏目管理」中不可见。
func (s *TemplateService) DeleteTemplate(id uint) error {
	var columns []models.Column
	if err := utils.DB.Where("template_id = ?", id).Order("id").Find(&columns).Error; err != nil {
		return err
	}
	switch len(columns) {
	case 0:
	case 1:
		return fmt.Errorf("该模板下仍有栏目「%s」，请先删除或调整该栏目后再删除模板", columns[0].Name)
	default:
		return fmt.Errorf("该模板下仍有 %d 个栏目，请先删除或调整这些栏目后再删除模板", len(columns))
	}
	return utils.DB.Unscoped().Delete(&models.Template{}, id).Error
}
