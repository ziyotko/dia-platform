package models

import "time"

// OperationLog records admin write operations for audit purposes (操作日志).
type OperationLog struct {
	ID          uint64    `gorm:"primarykey" json:"id"`
	MemberID    uint64    `gorm:"index;default:0" json:"member_id"` // 操作人会员 ID
	Username    string    `gorm:"size:64;index" json:"username"`    // 操作人用户名
	Module      string    `gorm:"size:128" json:"module"`           // 路由模块（FullPath）
	Action      string    `gorm:"size:255" json:"action"`           // 方法 + 路径
	Method      string    `gorm:"size:10" json:"method"`
	Path        string    `gorm:"size:255;index" json:"path"`
	IP          string    `gorm:"size:64" json:"ip"`
	Params      string    `gorm:"type:text" json:"params"`   // 请求参数（截断存储）
	Result      string    `gorm:"type:text" json:"result"`   // 响应结果（截断存储）
	Status      int       `gorm:"default:1" json:"status"`   // 1=成功 0=失败
	Duration    int64     `gorm:"default:0" json:"duration"` // 耗时(ms)
	OperationAt time.Time `json:"operation_at"`
}

func (OperationLog) TableName() string {
	return "member_operation_logs"
}
