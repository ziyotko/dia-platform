package services

import (
	"server/models"
	"server/utils"
)

type StaticPageService struct{}

// GetStaticTemplates 获取静态化页面列表。
// 页面层已合并进模板（原 page 表已移除），因此这里直接返回启用中的模板：
// pageType 为空表示全部类型，否则按 type（home/column/detail/special）过滤。
func (s *StaticPageService) GetStaticTemplates(pageType string) ([]models.Template, error) {
	var templates []models.Template
	query := utils.DB.Model(&models.Template{})
	if pageType != "" {
		query = query.Where("type = ?", pageType)
	}
	query = query.Where("status = ?", 1)
	err := query.Order("id DESC").Find(&templates).Error
	return templates, err
}
