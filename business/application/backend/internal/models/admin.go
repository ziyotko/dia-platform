package models

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
