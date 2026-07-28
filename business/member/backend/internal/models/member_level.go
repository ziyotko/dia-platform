package models

// MemberLevel represents a predefined member rank/grade
type MemberLevel struct {
	BaseModel
	Name        string `gorm:"size:64;not null" json:"name"`
	Level       int    `gorm:"not null;default:0" json:"level"` // ordering, lower = lower rank
	Description string `gorm:"size:255" json:"description"`
}

func (MemberLevel) TableName() string {
	return "member_levels"
}
