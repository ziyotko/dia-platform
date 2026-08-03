package models

// AuditLog records all operations (immutable)
type AuditLog struct {
	BaseModel
	UserID     uint64 `gorm:"index" json:"userId"`
	Username   string `gorm:"size:64" json:"username"`
	UserType   string `gorm:"size:16" json:"userType"` // member/admin
	Action     string `gorm:"size:64;index" json:"action"`
	Resource   string `gorm:"size:64" json:"resource"`
	ResourceID uint64 `gorm:"default:0" json:"resourceId"`
	Detail     string `gorm:"type:text" json:"detail"` // JSON detail
	IP         string `gorm:"size:64" json:"ip"`
	UserAgent  string `gorm:"size:512" json:"userAgent"`
}

func (AuditLog) TableName() string {
	return "conference_audit_logs"
}
