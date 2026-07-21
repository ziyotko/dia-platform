package models

type Setting struct {
	BaseModel
	Category string `gorm:"size:64;index;comment:分类" json:"category"`
	Key      string `gorm:"size:128;index;comment:配置键" json:"key"`
	Value    string `gorm:"type:text;comment:配置值" json:"value"`
	DefaultValue string `gorm:"type:text;comment:默认值" json:"defaultValue"`
	Type     string `gorm:"size:32;comment:值类型 string/json/number/boolean" json:"type"`
	Remark   string `gorm:"size:512;comment:备注" json:"remark"`
}

func (Setting) TableName() string {
	return "base_setting"
}
