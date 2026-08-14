package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

type Link struct {
	gorm.Model
	Name        string `gorm:"size:200;not null" json:"name"`
	Url         string `gorm:"size:500;not null" json:"url"`
	Logo        string `gorm:"size:500" json:"logo"`
	Description string `gorm:"size:500" json:"description"`
	PageID      uint   `gorm:"not null;index" json:"pageId"`
	ColumnID    uint   `gorm:"default:0;index" json:"columnId"`
	Sort        int    `gorm:"default:0" json:"sort"`
	Status      int    `gorm:"default:1;index" json:"status"`
	Author      string `gorm:"size:100" json:"author"`
	AuthorCode  string `gorm:"size:100" json:"authorCode"`
}

func (l Link) MarshalJSON() ([]byte, error) {
	type Alias Link
	return json.Marshal(&struct {
		CreatedAt string `json:"createTime"`
		UpdatedAt string `json:"updatedAt"`
		Alias
	}{
		CreatedAt: l.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: l.UpdatedAt.Format("2006-01-02 15:04:05"),
		Alias:     (Alias)(l),
	})
}
