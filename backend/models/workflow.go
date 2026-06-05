package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Workflow struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"createTime"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Name        string         `gorm:"size:200;not null" json:"name"`
	Status      int            `gorm:"default:1" json:"status"` // 1启用 0禁用
	Description string         `gorm:"size:500" json:"description"`
	Nodes       []WorkflowNode `gorm:"foreignKey:WorkflowID;references:ID;constraint:OnDelete:CASCADE;" json:"nodes,omitempty"`
}

func (w Workflow) MarshalJSON() ([]byte, error) {
	type Alias Workflow
	return json.Marshal(&struct {
		CreatedAt string `json:"createTime"`
		UpdatedAt string `json:"updatedAt"`
		*Alias
	}{
		CreatedAt: w.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: w.UpdatedAt.Format("2006-01-02 15:04:05"),
		Alias:     (*Alias)(&w),
	})
}

type WorkflowNode struct {
	ID         uint `gorm:"primarykey" json:"id"`
	WorkflowID uint `gorm:"not null;index" json:"workflowId"`
	Name       string `gorm:"size:200;not null" json:"name"`
	ApproverID uint   `gorm:"index" json:"approverId"`
	SortOrder  int    `gorm:"default:0" json:"sortOrder"`
}
