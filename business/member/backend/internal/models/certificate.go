package models

// CertificateStatus constants
const (
	CertStatusActive  = "active"  // 有效
	CertStatusExpired = "expired" // 已过期
)

// Certificate represents a membership certificate
type Certificate struct {
	BaseModel
	MemberID       uint64     `gorm:"index;not null" json:"member_id"`
	Member         Member     `gorm:"foreignKey:MemberID" json:"member,omitempty"`
	CertNo         string     `gorm:"size:64;not null" json:"cert_no"`
	IssuedAt       *LocalTime `gorm:"type:datetime" json:"issued_at"`
	ExpireAt       *LocalTime `gorm:"type:datetime" json:"expire_at"`
	FilePath       string     `gorm:"size:255" json:"file_path"`
	Status         string     `gorm:"size:20;default:active" json:"status"`
	LevelID        uint64     `gorm:"default:0" json:"level_id"`
	LevelName      string     `gorm:"size:64" json:"level_name"`
	CertTemplateID uint64     `gorm:"default:0" json:"cert_template_id"`
}

func (Certificate) TableName() string {
	return "member_certificates"
}
