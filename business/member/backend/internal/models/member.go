package models

import "time"

// MemberStatus represents the member lifecycle status
const (
	MemberStatusRegistering   = "registering"     // 注册中
	MemberStatusPendingReview = "pending_review"  // 待审核
	MemberStatusPendingPay    = "pending_payment" // 待缴费
	MemberStatusActive        = "active"          // 正式会员
	MemberStatusRejected      = "rejected"        // 已拒绝
	MemberStatusExpired       = "expired"         // 已过期
)

// MemberType
const (
	MemberTypeUnit     = "unit"     // 单位会员
	MemberTypePersonal = "personal" // 个人会员
)

// Member represents a member user
type Member struct {
	BaseModel
	Username    string `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Password    string `gorm:"size:255;not null" json:"-"`
	Mobile      string `gorm:"size:20" json:"mobile"`
	Email       string `gorm:"size:128" json:"email"`
	MemberType  string `gorm:"size:20;default:unit" json:"member_type"`
	MemberLevel string `gorm:"size:20;default:''" json:"member_level"`
	Status      string `gorm:"size:20;default:registering" json:"status"`
	IsAdmin     bool   `gorm:"default:false" json:"is_admin"`
	Avatar      string `gorm:"size:255" json:"avatar"`

	// 主入会机构（非数据库字段，管理端查询时由 service 填充）：
	// 会员通过入会申请审批/缴费加入的机构名称。
	OrgName string `gorm:"-" json:"org_name"`

	// GeneratedPassword 管理端新增会员未填密码时生成的随机初始密码（非数据库字段）。
	// 只在创建响应里返回一次供管理员转告会员，其它接口不会带该字段。
	GeneratedPassword string `gorm:"-" json:"generated_password,omitempty"`

	// Company info (for unit members)
	CompanyName       string `gorm:"size:255" json:"company_name"`
	CreditCode        string `gorm:"size:64" json:"credit_code"`
	LegalPerson       string `gorm:"size:64" json:"legal_person"`
	ContactPerson     string `gorm:"size:64" json:"contact_person"`
	ContactTitle      string `gorm:"size:64" json:"contact_title"`
	ContactMobile     string `gorm:"size:20" json:"contact_mobile"`
	Industry          string `gorm:"size:64" json:"industry"`
	FoundedDate       string `gorm:"size:20" json:"founded_date"`
	RegisteredCapital string `gorm:"size:64" json:"registered_capital"`
	EmployeeCount     int    `gorm:"default:0" json:"employee_count"`
	BusinessScope     string `gorm:"type:text" json:"business_scope"`
	PostalCode        string `gorm:"size:16" json:"postal_code"`
	Address           string `gorm:"size:255" json:"address"`
	Website           string `gorm:"size:255" json:"website"`
	Description       string `gorm:"type:text" json:"description"`
	CertFile          string `gorm:"size:255" json:"cert_file"`

	// Personal info (for personal members)
	Name   string `gorm:"size:64" json:"name"`
	IDCard string `gorm:"size:32" json:"id_card"`

	LoginFailCount int        `gorm:"default:0" json:"-"`
	LockedUntil    *LocalTime `gorm:"type:datetime" json:"-"`

	// PasswordChangedAt 最后一次修改密码的时间（本人改密 / 管理员重置密码时写入，截断到秒）。
	// JWT 没有版本号，改密本身不会让已签发的 Token 失效；Auth 中间件会比较 Token 的 iat
	// 与本字段，使改密前签发的 Token 立即作废（原先最长要等 24h 自然过期）。
	PasswordChangedAt *LocalTime `gorm:"type:datetime" json:"-"`
	// TokenInvalidBefore 该时间之前签发的 Token 一律作废。由管理员触发的账号变更
	// （会籍状态变更等）写入，使旧 Token 立即失效，无需等 JWT 自然过期。
	TokenInvalidBefore *LocalTime `gorm:"type:datetime" json:"-"`
}

func (Member) TableName() string {
	return "member_users"
}

// NewInvalidBefore 返回「截断到秒」的当前时间，用于 password_changed_at / token_invalid_before。
// 必须截断：这两列是没有小数秒的 datetime，MySQL 写入时会四舍五入；若不截断，
// 进位后的失效点会晚于「同一秒内新登录签发」的 Token，导致改密后重新登录立刻被判失效。
func NewInvalidBefore() *LocalTime {
	return &LocalTime{Time: time.Now().Truncate(time.Second)}
}

// IsTokenStale 判断「签发时间早于失效点」的 Token 是否应作废。
// 两个时间同为秒级精度：恰好等于失效点（同一秒内签发）视为有效。
func IsTokenStale(issuedAt time.Time, invalidBefore *LocalTime) bool {
	return invalidBefore != nil && !invalidBefore.Time.IsZero() && issuedAt.Before(invalidBefore.Time)
}
