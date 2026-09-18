package models

// MemberCertificateTemplate represents a certificate style/template
type MemberCertificateTemplate struct {
	BaseModel
	Name         string      `gorm:"size:64;not null" json:"name"`
	LevelID      uint64      `gorm:"uniqueIndex:uk_cert_tpl_level;not null" json:"level_id"`
	Level        MemberLevel `gorm:"foreignKey:LevelID" json:"level,omitempty"`
	TemplateFile string      `gorm:"size:255" json:"template_file"` // PDF file path
	// FileExists 模板文件在服务器上是否真实存在（非数据库列）。
	// 历史数据里存在指向已丢失文件/旧路径格式的值，列表页据此提醒管理员重新上传。
	FileExists bool `gorm:"-" json:"file_exists"`
}

func (MemberCertificateTemplate) TableName() string {
	return "member_certificate_templates"
}
