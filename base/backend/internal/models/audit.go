package models

import "time"

type OperationLog struct {
	BaseModel
	TenantID    uint64    `gorm:"index;comment:租户ID" json:"tenantId"`
	UserID      uint64    `gorm:"index;comment:用户ID" json:"userId"`
	Username    string    `gorm:"size:64;comment:用户名" json:"username"`
	Module      string    `gorm:"size:64;comment:模块" json:"module"`
	Action      string    `gorm:"size:64;comment:操作" json:"action"`
	Method      string    `gorm:"size:16;comment:请求方法" json:"method"`
	Path        string    `gorm:"size:512;comment:请求路径" json:"path"`
	IP          string    `gorm:"size:64;comment:IP" json:"ip"`
	Params      string    `gorm:"type:text;comment:请求参数" json:"params"`
	Result      string    `gorm:"type:text;comment:响应结果" json:"result"`
	Status      int       `gorm:"comment:状态 1成功 0失败" json:"status"`
	Duration    int64     `gorm:"comment:耗时ms" json:"duration"`
	OperationAt time.Time `gorm:"index;comment:操作时间" json:"operationAt"`
}

func (OperationLog) TableName() string {
	return "base_operation_log"
}
