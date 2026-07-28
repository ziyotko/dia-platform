package models

// Organization represents an association branch/organization
type Organization struct {
	BaseModel
	Name        string `gorm:"size:255;not null" json:"name"`
	ParentID    uint64 `gorm:"default:0" json:"parent_id"`
	Type        string `gorm:"size:20;default:branch" json:"type"` // association/branch/committee
	Description string `gorm:"type:text" json:"description"`
	ContactInfo string `gorm:"size:255" json:"contact_info"`
	Sort        int    `gorm:"default:0" json:"sort"`

	Children []*Organization `gorm:"-" json:"children,omitempty"`
}

func (Organization) TableName() string {
	return "member_organizations"
}
