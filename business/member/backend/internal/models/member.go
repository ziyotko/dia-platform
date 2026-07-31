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

	// Company info (for unit members)
	CompanyName   string `gorm:"size:255" json:"company_name"`
	CreditCode    string `gorm:"size:64" json:"credit_code"`
	LegalPerson   string `gorm:"size:64" json:"legal_person"`
	ContactPerson string `gorm:"size:64" json:"contact_person"`
	Address       string `gorm:"size:255" json:"address"`
	Website       string `gorm:"size:255" json:"website"`
	Description   string `gorm:"type:text" json:"description"`
	CertFile      string `gorm:"size:255" json:"cert_file"`

	LoginFailCount int        `gorm:"default:0" json:"-"`
	LockedUntil    *LocalTime `gorm:"type:datetime" json:"-"`
}

func (Member) TableName() string {
	return "member_users"
}
