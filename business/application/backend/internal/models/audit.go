package models

// AuditLog records backend operations
type AuditLog struct {
	BaseModel
	AdminID uint64 `gorm:"index" json:"adminId"`
	Admin   string `gorm:"size:64" json:"admin"`
	Module  string `gorm:"size:64" json:"module"`
	Action  string `gorm:"size:128" json:"action"`
	IP      string `gorm:"size:64" json:"ip"`
	Detail  string `gorm:"type:text" json:"detail"`
}

func (AuditLog) TableName() string {
	return "application_audit_logs"
}
