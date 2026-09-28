package models

import "time"

// Admin represents a backend operator (管理人 / 评审人)
type Admin struct {
	BaseModel
	Username string `gorm:"size:64;uniqueIndex" json:"username"`
	Password string `gorm:"size:128" json:"-"`
	RealName string `gorm:"size:64" json:"realName"`
	Phone    string `gorm:"size:32" json:"phone"`
	Email    string `gorm:"size:128" json:"email"`
	RoleCode string `gorm:"size:32" json:"roleCode"`
	Status   int    `gorm:"default:1" json:"status"`
	// PasswordChangedAt 最后一次修改密码的时间（本人改密 / 管理员在账号管理里改密时写入）。
	// AdminAuth 比较 Token 的 iat 与本字段，使改密前签发的 Token 立即作废。
	PasswordChangedAt *time.Time `gorm:"type:datetime" json:"-"`
}

func (Admin) TableName() string {
	return "application_admins"
}

type Role struct {
	BaseModel
	Code        string `gorm:"size:32;uniqueIndex" json:"code"`
	Name        string `gorm:"size:64" json:"name"`
	Permissions string `gorm:"type:text" json:"permissions"` // JSON array of permission codes
	Description string `gorm:"size:256" json:"description"`
}

func (Role) TableName() string {
	return "application_roles"
}
