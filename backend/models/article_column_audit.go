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
	ArticleID       uint           `gorm:"not null;index:idx_aca_article_column;index:idx_aca_article_status" json:"articleId"`
	ColumnID        uint           `gorm:"not null;index:idx_aca_article_column" json:"columnId"`
	WorkflowID      uint           `gorm:"not null;index" json:"workflowId"`
	CurrentNodeID   uint           `gorm:"default:0;index:idx_aca_status_current_node" json:"currentNodeId"`
	Status          int            `gorm:"default:0;index:idx_aca_article_status;index:idx_aca_status_current_node" json:"status"` // 0进行中 1已通过 2已驳回
	ApproveRemark   string         `gorm:"size:500" json:"approveRemark"`
	RejectRemark    string         `gorm:"size:500" json:"rejectRemark"`
	ApproveUserID   uint           `gorm:"index" json:"approveUserId"`
	ApproveUserName string         `gorm:"size:50" json:"approveUserName"`
	ApproveTime     *time.Time     `json:"approveTime"`
}
