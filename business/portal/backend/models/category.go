package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

type Category struct {
	gorm.Model
	Name        string `gorm:"size:100;not null" json:"name"`
	Code        string `gorm:"size:100;not null" json:"code"`
	Description string `gorm:"size:500" json:"description"`
	Sort        int    `gorm:"default:0" json:"sort"`
	Status      int    `gorm:"default:1;index" json:"status"`
}

func (c Category) MarshalJSON() ([]byte, error) {
	type Alias Category
	return json.Marshal(&struct {
		CreatedAt string `json:"createTime"`
		UpdatedAt string `json:"updatedAt"`
		*Alias
	}{
		CreatedAt: c.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: c.UpdatedAt.Format("2006-01-02 15:04:05"),
		Alias:     (*Alias)(&c),
	})
}
