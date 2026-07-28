package models

// PasswordReset represents a password reset token
type PasswordReset struct {
	BaseModel
	MemberID uint64     `gorm:"index;not null" json:"member_id"`
	Token    string     `gorm:"size:128;uniqueIndex;not null" json:"token"`
	ExpireAt *LocalTime `gorm:"type:datetime;not null" json:"expire_at"`
	Used     bool       `gorm:"default:false" json:"used"`
}

func (PasswordReset) TableName() string {
	return "member_password_resets"
}
