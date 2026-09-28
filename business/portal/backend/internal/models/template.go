package models

import (
	"encoding/json"
	"time"
)

// Template 页面模板。模板即"页面"：原 `page` 中间层已合并到本表（1:1 关系取消）。
// Type 即页面类型（home/column/detail/special）；Code/RoutePath 来自原 page 表，用于路由拼接与静态化。
type Template struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Code        string    `gorm:"size:100;index" json:"code"`
	Type        string    `gorm:"size:20;not null;index" json:"type"`
	RoutePath   string    `gorm:"size:200" json:"routePath"`
	Description string    `gorm:"size:500" json:"description"`
	Status      int       `gorm:"default:1;index" json:"status"`
	SourceCode  string    `gorm:"type:text" json:"sourceCode"`
	Layout      string    `gorm:"type:text" json:"layout"`
}

// MarshalJSON 把时间字段统一输出为 `2006-01-02 15:04:05`（与已移除的 Page 模型一致，避免前端显示带 T 的 ISO 时间）。
func (t Template) MarshalJSON() ([]byte, error) {
	type Alias Template
	return json.Marshal(&struct {
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
		Alias
	}{
		CreatedAt: t.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: t.UpdatedAt.Format("2006-01-02 15:04:05"),
		Alias:     (Alias)(t),
	})
}
