package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

type Page struct {
	gorm.Model
	Name        string `gorm:"size:100;not null" json:"name"`
	Code        string `gorm:"size:100;not null" json:"code"`
	PageType    string `gorm:"size:20;not null;index" json:"pageType"`
	RoutePath   string `gorm:"size:200" json:"routePath"`
	TemplateID  uint   `gorm:"default:0;index" json:"templateId"`
	Template    string `gorm:"size:100" json:"template"`
	Description string `gorm:"size:500" json:"description"`
	Status      int    `gorm:"default:1;index" json:"status"`
}

func (p Page) MarshalJSON() ([]byte, error) {
	type Alias Page
	return json.Marshal(&struct {
		CreatedAt string `json:"createTime"`
		UpdatedAt string `json:"updatedAt"`
		Alias
	}{
		CreatedAt: p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: p.UpdatedAt.Format("2006-01-02 15:04:05"),
		Alias:     (Alias)(p),
	})
}
