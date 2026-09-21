package models

import (
	"encoding/json"
	"time"
)

type Workflow struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	Name        string         `gorm:"size:200;not null" json:"name"`
	Status      int            `gorm:"index" json:"status"` // 1启用 0禁用
	Description string         `gorm:"size:500" json:"description"`
	Nodes       []WorkflowNode `gorm:"foreignKey:WorkflowID;references:ID;constraint:OnDelete:CASCADE;" json:"nodes,omitempty"`
}

func (w Workflow) MarshalJSON() ([]byte, error) {
	type Alias Workflow
	return json.Marshal(&struct {
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
		*Alias
	}{
		CreatedAt: w.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: w.UpdatedAt.Format("2006-01-02 15:04:05"),
		Alias:     (*Alias)(&w),
	})
}

type WorkflowNode struct {
	ID           uint   `gorm:"primarykey" json:"id"`
	WorkflowID   uint   `gorm:"not null;index:idx_workflow_node_workflow_sort" json:"workflowId"`
	Name         string `gorm:"size:200;not null" json:"name"`
	ApproverType string `gorm:"size:20;default:'user'" json:"approverType"`
	ApproverID   uint   `gorm:"index" json:"approverId"`
	SortOrder    int    `gorm:"default:0;index:idx_workflow_node_workflow_sort" json:"sortOrder"`
}
