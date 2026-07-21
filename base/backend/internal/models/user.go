package models

type User struct {
	BaseModel
	TenantID uint64 `gorm:"index;comment:租户ID" json:"tenantId"`
	Username string `gorm:"size:64;index;comment:用户名" json:"username"`
	Password string `gorm:"size:128;comment:密码" json:"-"`
	RealName string `gorm:"size:64;comment:真实姓名" json:"realName"`
	Phone    string `gorm:"size:32;comment:手机号" json:"phone"`
	Email    string `gorm:"size:128;comment:邮箱" json:"email"`
	Avatar   string `gorm:"size:512;comment:头像" json:"avatar"`
	Status   int    `gorm:"default:1;comment:状态 1启用 0禁用" json:"status"`
	IsAdmin  bool   `gorm:"default:false;comment:是否租户管理员" json:"isAdmin"`
	Roles    []Role `gorm:"many2many:base_user_role;" json:"roles,omitempty"`
}

func (User) TableName() string {
	return "base_user"
}
