package services

import (
	"fmt"

	"server/models"
	"server/utils"
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

func (s *TemplateService) CreateTemplate(template *models.Template) error {
	return utils.DB.Create(template).Error
}

func (s *TemplateService) UpdateTemplate(id uint, updates map[string]any) error {
	return utils.DB.Model(&models.Template{}).Where("id = ?", id).Updates(updates).Error
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
		return fmt.Errorf("该模板已被页面「%s」应用，请先在页面管理中解除绑定后再删除", pages[0].Name)
	default:
		return fmt.Errorf("该模板已被 %d 个页面应用，请先解除绑定后再删除", len(pages))
	}
	return utils.DB.Unscoped().Delete(&models.Template{}, id).Error
}
