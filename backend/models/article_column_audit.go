package models

import (
	"time"

	"gorm.io/gorm"
)

type ArticleColumnAudit struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time      `json:"createTime"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	ArticleID       uint           `gorm:"not null;index" json:"articleId"`
	ColumnID        uint           `gorm:"not null;index" json:"columnId"`
	WorkflowID      uint           `gorm:"not null;index" json:"workflowId"`
	CurrentNodeID   uint           `gorm:"default:0;index" json:"currentNodeId"`
	Status          int            `gorm:"default:0" json:"status"` // 0进行中 1已通过 2已驳回
	ApproveRemark   string         `gorm:"size:500" json:"approveRemark"`
	RejectRemark    string         `gorm:"size:500" json:"rejectRemark"`
	ApproveUserID   uint           `gorm:"index" json:"approveUserId"`
	ApproveUserName string         `gorm:"size:50" json:"approveUserName"`
	ApproveTime     *time.Time     `json:"approveTime"`
}
