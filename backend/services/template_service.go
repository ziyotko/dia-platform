package services

import (
	"server/models"
	"server/utils"
)

type TemplateService struct{}

type TemplateListResult struct {
	Total int64             `json:"total"`
	List  []models.Template `json:"list"`
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

	offset := (page - 1) * pageSize
	err = query.Order("type,id DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, err
	}

	return &TemplateListResult{
		Total: total,
		List:  list,
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

func (s *TemplateService) UpdateTemplate(id uint, updates map[string]interface{}) error {
	return utils.DB.Model(&models.Template{}).Where("id = ?", id).Updates(updates).Error
}

func (s *TemplateService) DeleteTemplate(id uint) error {
	return utils.DB.Unscoped().Delete(&models.Template{}, id).Error
}
