package models

// User represents an applicant (申报人) on the frontend side
type User struct {
	BaseModel
	Username     string `gorm:"size:64;uniqueIndex" json:"username"`
	Password     string `gorm:"size:128" json:"-"`
	RealName     string `gorm:"size:64" json:"realName"`
	Phone        string `gorm:"size:32" json:"phone"`
	Email        string `gorm:"size:128" json:"email"`
	IDCard       string `gorm:"size:32" json:"idCard"`
	Organization string `gorm:"size:256" json:"organization"`
	Position     string `gorm:"size:64" json:"position"`
	Status       int    `gorm:"default:1" json:"status"` // 1=enabled, 0=disabled
}

func (User) TableName() string {
	return "application_users"
}
