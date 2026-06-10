package models

import "time"

type ArticleColumnAuditHistory struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time `json:"createTime"`
	ArticleID    uint      `gorm:"not null;index" json:"articleId"`
	ColumnID     uint      `gorm:"not null;index" json:"columnId"`
	WorkflowID   uint      `gorm:"not null;index" json:"workflowId"`
	NodeID       uint      `gorm:"not null;index" json:"nodeId"`
	NodeName     string    `gorm:"size:200" json:"nodeName"`
	Action       int       `gorm:"default:1" json:"action"` // 1通过 2驳回
	OperatorID   uint      `gorm:"index" json:"operatorId"`
	OperatorName string    `gorm:"size:50" json:"operatorName"`
	Remark       string    `gorm:"size:500" json:"remark"`
}
