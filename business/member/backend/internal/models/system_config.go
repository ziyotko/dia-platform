package models

// SystemConfig represents key-value system configuration
type SystemConfig struct {
	BaseModel
	Key         string `gorm:"uniqueIndex;size:128;not null" json:"key"`
	Value       string `gorm:"type:text" json:"value"`
	Description string `gorm:"size:255" json:"description"`
}

func (SystemConfig) TableName() string {
	return "member_system_configs"
}
