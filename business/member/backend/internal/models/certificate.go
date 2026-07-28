package models

// Certificate represents a membership certificate
type Certificate struct {
	BaseModel
	MemberID uint64     `gorm:"index;not null" json:"member_id"`
	Member   Member     `gorm:"foreignKey:MemberID" json:"member,omitempty"`
	CertNo   string     `gorm:"size:64;not null" json:"cert_no"`
	IssuedAt *LocalTime `gorm:"type:datetime" json:"issued_at"`
	ExpireAt *LocalTime `gorm:"type:datetime" json:"expire_at"`
	FilePath string     `gorm:"size:255" json:"file_path"`
	Status   string     `gorm:"size:20;default:active" json:"status"`
}

func (Certificate) TableName() string {
	return "member_certificates"
}
