package models

import (
	"time"

	"gorm.io/gorm"
)

// ProfileChange records a member profile modification (资料变更记录).
// Each row stores one changed field: Name is the Chinese title of the field,
// OldContent is the value before the change and NewContent the value after.
// Records are append-only (CreatedAt is the change time); no updated_at.
// Passwords are never stored.
type ProfileChange struct {
	ID         uint64         `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time      `json:"created_at"` // 变更时间
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	MemberID   uint64         `gorm:"index;not null" json:"member_id"`
	Username   string         `gorm:"size:64;index" json:"username"` // 用户名
	Name       string         `gorm:"size:64" json:"name"`           // 修改内容对应的中文标题
	OldContent string         `gorm:"type:text" json:"old_content"`  // 修改前的内容
	NewContent string         `gorm:"type:text" json:"new_content"`  // 修改后的内容
	Operator   string         `gorm:"size:64" json:"operator"`       // 修改人
}

func (ProfileChange) TableName() string {
	return "member_profile_changes"
}
