package services

import (
	"server/models"
	"server/utils"
)

type TemplateService struct{}

type TemplateListResult struct {
	Total      int64             `json:"total"`
	List       []models.Template `json:"list"`
	PageCounts map[uint]int      `json:"-"`
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

	// 实时统计每个模板关联的页面数
	pageCounts := make(map[uint]int)
	if len(list) > 0 {
		ids := make([]uint, len(list))
		for i, t := range list {
			ids[i] = t.ID
		}

		type pageCountRow struct {
			TemplateID uint
			Count      int64
		}
		var rows []pageCountRow
		if err := utils.DB.Model(&models.Page{}).
			Select("template_id, COUNT(*) AS count").
			Where("template_id IN ?", ids).
			Group("template_id").
			Find(&rows).Error; err == nil {
			for _, r := range rows {
				pageCounts[r.TemplateID] = int(r.Count)
			}
		}
	}

	return &TemplateListResult{
		Total:      total,
		List:       list,
		PageCounts: pageCounts,
	}, nil
}

func (s *TemplateService) GetTemplateByID(id uint) (*models.Template, error) {
	var template models.Template
	if err := utils.DB.First(&template, id).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (s *TemplateService) CreateTemplate(template *models.Template) error {
	return utils.DB.Create(template).Error
}

func (s *TemplateService) UpdateTemplate(id uint, updates map[string]any) error {
	return utils.DB.Model(&models.Template{}).Where("id = ?", id).Updates(updates).Error
}

func (s *TemplateService) DeleteTemplate(id uint) error {
	return utils.DB.Unscoped().Delete(&models.Template{}, id).Error
}
