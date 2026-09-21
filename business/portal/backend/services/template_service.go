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
	Total      int64             `json:"total"`
	List       []models.Template `json:"list"`
	PageCounts map[uint]int      `json:"-"`
	PageNames  map[uint]string   `json:"-"`
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

	// 实时统计每个模板关联的页面：一个模板只能应用一个页面（services.PageService 已有约束），
	// 这里统计出数量与应用页面名，供删除确认等提示使用（兼容历史数据中可能存在的多条绑定）。
	pageCounts := make(map[uint]int)
	pageNames := make(map[uint]string)
	if len(list) > 0 {
		ids := make([]uint, len(list))
		for i, t := range list {
			ids[i] = t.ID
		}

		var pages []models.Page
		if err := utils.DB.Model(&models.Page{}).
			Select("template_id, name").
			Where("template_id IN ?", ids).
			Order("id").
			Find(&pages).Error; err == nil {
			for _, p := range pages {
				pageCounts[p.TemplateID]++
				if pageNames[p.TemplateID] == "" {
					pageNames[p.TemplateID] = p.Name
				}
			}
		}
	}

	return &TemplateListResult{
		Total:      total,
		List:       list,
		PageCounts: pageCounts,
		PageNames:  pageNames,
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
// 与「一个模板只能应用一个页面」配套：仍被页面绑定时拒绝删除，
// 避免页面留下失效的 template_id（页面会显示为空模板且无从修复）。
func (s *TemplateService) DeleteTemplate(id uint) error {
	var pages []models.Page
	if err := utils.DB.Where("template_id = ?", id).Order("id").Find(&pages).Error; err != nil {
		return err
	}
	switch len(pages) {
	case 0:
	case 1:
		return fmt.Errorf("该模板已被页面「%s」应用，请先解除页面绑定后再删除", pages[0].Name)
	default:
		return fmt.Errorf("该模板已被 %d 个页面应用，请先解除绑定后再删除", len(pages))
	}
	return utils.DB.Unscoped().Delete(&models.Template{}, id).Error
}
