package models

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
}

func (Member) TableName() string {
	return "member_users"
}
