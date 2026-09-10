package models

// ProfileChange records a member profile modification (资料变更记录).
// Each row stores one modification: the member name and the full tracked
// profile content before and after the change. Passwords are never stored.
type ProfileChange struct {
	BaseModel
	MemberID   uint64     `gorm:"index;not null" json:"member_id"`
	ChangedAt  *LocalTime `gorm:"type:datetime" json:"changed_at"` // 变更时间
	OldName    string     `gorm:"size:255" json:"old_name"`        // 原始name
	OldContent string     `gorm:"type:text" json:"old_content"`    // 原始内容
	NewName    string     `gorm:"size:255" json:"new_name"`        // 变更name
	NewContent string     `gorm:"type:text" json:"new_content"`    // 变更后内容
	Operator   string     `gorm:"size:64" json:"operator"`         // 修改人
}

func (ProfileChange) TableName() string {
	return "member_profile_changes"
}
