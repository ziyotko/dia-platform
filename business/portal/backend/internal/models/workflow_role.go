package models

import (
	"time"
)

// WorkflowRole 流程角色
type WorkflowRole struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Name        string    `gorm:"size:50;not null" json:"name"`
	Code        string    `gorm:"unique;size:50;not null" json:"code"`
	Description string    `gorm:"size:255" json:"description"`
	Status      int       `gorm:"index" json:"status"`
}

// WorkflowRoleUser 流程角色与用户关联
type WorkflowRoleUser struct {
	ID             uint `gorm:"primarykey" json:"id"`
	WorkflowRoleID uint `gorm:"not null;index" json:"workflowRoleId"`
	UserID         uint `gorm:"not null;index" json:"userId"`
}
