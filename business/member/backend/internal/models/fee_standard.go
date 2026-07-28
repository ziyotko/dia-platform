package models

// FeeStandard defines the annual fee amount for a specific member level.
// Multiple years can be configured for each level.
type MemberFeeStandard struct {
	BaseModel
	LevelID uint64      `gorm:"uniqueIndex:idx_level_year;not null" json:"level_id"`
	Year    int         `gorm:"uniqueIndex:idx_level_year;not null" json:"year"`
	Amount  float64     `gorm:"type:decimal(10,2);not null;default:0" json:"amount"`
	Level   MemberLevel `gorm:"foreignKey:LevelID" json:"level,omitempty"`
}

func (MemberFeeStandard) TableName() string {
	return "member_fee_standards"
}
