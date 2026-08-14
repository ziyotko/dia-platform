package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

type Column struct {
	gorm.Model
	Name        string   `gorm:"size:100;not null" json:"name"`
	Code        string   `gorm:"size:100;not null" json:"code"`
	PageID      uint     `gorm:"not null;index" json:"pageId"`
	ParentID    uint     `gorm:"default:0;index" json:"parentId"`
	RoutePath   string   `gorm:"size:200" json:"routePath"`
	Template    string   `gorm:"size:100" json:"template"`
	Description string   `gorm:"size:500" json:"description"`
	Sort        int      `gorm:"default:0" json:"sort"`
	Status      int      `gorm:"default:1" json:"status"`
	DisplayType int      `gorm:"default:1" json:"displayType"`
	WorkflowID  *uint    `gorm:"index" json:"workflowId"`
	Workflow    Workflow `gorm:"foreignKey:WorkflowID;references:ID" json:"workflow"`
}

func (c Column) MarshalJSON() ([]byte, error) {
	type Alias Column
	return json.Marshal(&struct {
		CreatedAt string `json:"createTime"`
		UpdatedAt string `json:"updatedAt"`
		Alias
	}{
		CreatedAt: c.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: c.UpdatedAt.Format("2006-01-02 15:04:05"),
		Alias:     (Alias)(c),
	})
}
