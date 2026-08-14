package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

type Tag struct {
	gorm.Model
	Name   string `gorm:"size:100;not null" json:"name"`
	Color  string `gorm:"size:20;default:'rgb(64,158,255)'" json:"color"`
	Status int    `gorm:"default:1;index" json:"status"`
}

func (t Tag) MarshalJSON() ([]byte, error) {
	type Alias Tag
	return json.Marshal(&struct {
		CreatedAt string `json:"createTime"`
		UpdatedAt string `json:"updatedAt"`
		*Alias
	}{
		CreatedAt: t.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: t.UpdatedAt.Format("2006-01-02 15:04:05"),
		Alias:     (*Alias)(&t),
	})
}
