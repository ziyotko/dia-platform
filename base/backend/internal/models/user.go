package models

import "time"

type User struct {
	BaseModel
	// 用户名唯一性由数据库唯一索引 (tenant_id, username) 保证：同租户内不能重名，跨租户可同名。
	// 只靠 service 层校验在并发下会漏（两个请求同时通过校验 → 两条同名记录），必须落索引兜底。
	TenantID uint64 `gorm:"index;uniqueIndex:uk_base_user_tenant_username;comment:租户ID" json:"tenantId"`
	Username string `gorm:"size:64;uniqueIndex:uk_base_user_tenant_username;comment:用户名" json:"username"`
	Password string `gorm:"size:128;comment:密码" json:"-"`
	RealName string `gorm:"size:64;comment:真实姓名" json:"realName"`
	Phone    string `gorm:"size:32;comment:手机号" json:"phone"`
	Email    string `gorm:"size:128;comment:邮箱" json:"email"`
	Avatar   string `gorm:"size:512;comment:头像" json:"avatar"`
	Status   int    `gorm:"default:1;comment:状态 1启用 0禁用" json:"status"`
	IsAdmin  bool   `gorm:"default:false;comment:是否租户管理员" json:"isAdmin"`
	// OrganizationID 所属机构（0 表示未分配）；机构树见 base_organization
	OrganizationID uint64 `gorm:"index;comment:所属机构ID" json:"organizationId"`
	Roles          []Role `gorm:"many2many:base_user_role;" json:"roles,omitempty"`
	// PasswordChangedAt 最后一次修改密码的时间（本人改密 / 管理员重置或修改密码时写入，秒级）。
	// JWT 没有版本号，改密本身不会让已签发的 Token 失效；JWTAuth 会比较 iat 与本字段，
	// 使改密前签发的 Token 立即作废（且不依赖 Redis：Redis 数据丢失也不会漏）。
	PasswordChangedAt *time.Time `gorm:"comment:最后改密时间" json:"-"`
}

func (User) TableName() string {
	return "base_user"
}
