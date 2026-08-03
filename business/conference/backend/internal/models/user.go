package models

type User struct {
	BaseModel
	Username    string `gorm:"size:64;uniqueIndex" json:"username"`
	Password    string `gorm:"size:128" json:"-"`
	RealName    string `gorm:"size:64" json:"realName"`
	Phone       string `gorm:"size:32" json:"phone"`
	Email       string `gorm:"size:128" json:"email"`
	Avatar      string `gorm:"size:512" json:"avatar"`
	MemberLevel string `gorm:"size:64" json:"memberLevel"`
	Company     string `gorm:"size:128" json:"company"`
	Branch      string `gorm:"size:128" json:"branch"`
	Position    string `gorm:"size:64" json:"position"`
	IsValid     bool   `gorm:"default:true" json:"isValid"`
	Status      int    `gorm:"default:1" json:"status"` // 1=enabled, 0=disabled
}

func (User) TableName() string {
	return "conference_users"
}
