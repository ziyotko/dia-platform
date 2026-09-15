package models

// 系统生成的会籍变更原因常量。
// 管理员手工变更等级时原因由管理员填写（不受这些常量限制），
// 其余由系统写入的记录统一使用下列常量，避免字面量散落各处。
const (
	ReasonMemberCreate  = "新增会员" // 管理员新增会员 / 审批通过入会
	ReasonFeePaid       = "缴费确认" // 费用记录首次变为已缴费
	ReasonLeaveOrg      = "退出机构" // 会员主动退出某机构
	ReasonMemberExpired = "会员到期" // 管理员将会员置为已过期
	ReasonMemberResume  = "恢复会籍" // 会员由已过期恢复为正式会员
	ReasonCertRenew     = "证书续期" // 会员自助续证（年度延续）
)

// MemberLevelChange records a membership level change (会籍变更记录).
// CreatedAt is treated as the change time (变更时间).
type MemberLevelChange struct {
	BaseModel
	MemberID     uint64 `gorm:"index;not null" json:"member_id"`
	Username     string `gorm:"size:64" json:"username"`       // 变更用户名
	MemberName   string `gorm:"size:255" json:"member_name"`   // 公司名称或个人姓名
	MemberType   string `gorm:"size:20" json:"member_type"`    // 类型: unit / personal
	ChangeYear   int    `gorm:"default:0" json:"change_year"`  // 变更年份
	OrgID        uint64 `gorm:"default:0" json:"org_id"`       // 入会机构 ID
	OrgName      string `gorm:"size:255" json:"org_name"`      // 入会机构
	OldLevelID   uint64 `gorm:"default:0" json:"old_level_id"` // 原始会籍 ID
	OldLevelName string `gorm:"size:64" json:"old_level_name"` // 原始会籍
	NewLevelID   uint64 `gorm:"default:0" json:"new_level_id"` // 新的会籍 ID
	NewLevelName string `gorm:"size:64" json:"new_level_name"` // 新的会籍
	Reason       string `gorm:"size:500" json:"reason"`        // 变更原因
	Operator     string `gorm:"size:64" json:"operator"`       // 变更人
}

func (MemberLevelChange) TableName() string {
	return "member_level_changes"
}
