package models

type LoginLog struct {
	BaseModel
	TenantID uint64 `gorm:"index;comment:租户ID" json:"tenantId"`
	UserID   uint64 `gorm:"index;comment:用户ID" json:"userId"`
	Username string `gorm:"size:64;comment:用户名" json:"username"`
	IP       string `gorm:"size:64;comment:IP" json:"ip"`
	Agent    string `gorm:"size:512;comment:UserAgent" json:"agent"`
	Status   int    `gorm:"comment:状态 1成功 0失败" json:"status"`
	Message  string `gorm:"size:256;comment:消息" json:"message"`
}

func (LoginLog) TableName() string {
	return "base_login_log"
}
