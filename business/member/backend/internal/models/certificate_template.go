package models

// CertificateTemplate represents a certificate style/template
type CertificateTemplate struct {
	BaseModel
	Name         string      `gorm:"size:64;not null" json:"name"`
	LevelID      uint64      `gorm:"index;not null" json:"level_id"`
	Level        MemberLevel `gorm:"foreignKey:LevelID" json:"level,omitempty"`
	TemplateFile string      `gorm:"size:255" json:"template_file"` // PDF file path
}

func (CertificateTemplate) TableName() string {
	return "certificate_templates"
}
