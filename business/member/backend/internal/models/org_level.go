package models

// OrgLevel represents the association between an organization and a member level
type OrgLevel struct {
	BaseModel
	OrgID   uint64       `gorm:"uniqueIndex:idx_org_level;not null" json:"org_id"`
	LevelID uint64       `gorm:"uniqueIndex:idx_org_level;not null" json:"level_id"`
	Org     Organization `gorm:"foreignKey:OrgID" json:"org,omitempty"`
	Level   MemberLevel  `gorm:"foreignKey:LevelID" json:"level,omitempty"`
}

func (OrgLevel) TableName() string {
	return "org_levels"
}
