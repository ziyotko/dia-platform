package models

// MemberCertificateTemplate represents a certificate style/template
type MemberCertificateTemplate struct {
	BaseModel
	Name         string      `gorm:"size:64;not null" json:"name"`
	LevelID      uint64      `gorm:"uniqueIndex;not null" json:"level_id"`
	Level        MemberLevel `gorm:"foreignKey:LevelID" json:"level,omitempty"`
	TemplateFile string      `gorm:"size:255" json:"template_file"` // PDF file path
}

func (MemberCertificateTemplate) TableName() string {
	return "member_certificate_templates"
}
