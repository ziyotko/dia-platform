package services

import (
	"server/models"
	"server/utils"
)

type StaticPageService struct{}

// StaticTemplateItem 静态化页面列表项（首页 / 专题页）：只下发静态化页面前端所需的字段
// （不含 sourceCode/layout 两个大字段），并额外提供 url——可直接访问的地址，供前端「预览」新窗口打开。
type StaticTemplateItem struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	Type      string `json:"type"`
	RoutePath string `json:"routePath"`
	Status    int    `json:"status"`
	URL       string `json:"url"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// GetStaticTemplates 获取静态化页面列表。
// 页面层已合并进模板（原 page 表已移除），因此这里直接返回启用中的模板：
// pageType 为空表示全部类型，否则按 type（home/column/detail/special）过滤。
func (s *StaticPageService) GetStaticTemplates(pageType string) ([]StaticTemplateItem, error) {
	var templates []models.Template
	query := utils.DB.Model(&models.Template{})
	if pageType != "" {
		query = query.Where("type = ?", pageType)
	}
	query = query.Where("status = ?", 1)
	if err := query.Order("id DESC").Find(&templates).Error; err != nil {
		return nil, err
	}

	baseURL := SiteBaseURL()
	items := make([]StaticTemplateItem, 0, len(templates))
	for _, t := range templates {
		items = append(items, StaticTemplateItem{
			ID:        t.ID,
			Name:      t.Name,
			Code:      t.Code,
			Type:      t.Type,
			RoutePath: t.RoutePath,
			Status:    t.Status,
			URL:       BuildPageAccessURL(baseURL, t.RoutePath),
			CreatedAt: t.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: t.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return items, nil
}
