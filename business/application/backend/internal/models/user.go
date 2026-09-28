package models

import "time"

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
	// PasswordChangedAt 最后一次修改密码的时间（本人改密 / 管理员在账号管理里改密时写入）。
	// JWT 没有版本号，改密本身不会让已签发的 Token 失效；UserAuth 会比较 Token 的 iat
	// 与本字段，使改密前签发的 Token 立即作废（原先最长要等 24h 自然过期）。
	PasswordChangedAt *time.Time `gorm:"type:datetime" json:"-"`
}

func (User) TableName() string {
	return "application_users"
}
