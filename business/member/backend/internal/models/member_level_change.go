package models

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
