package models

// Organization represents an association branch/organization
type Organization struct {
	BaseModel
	Name     string `gorm:"size:255;not null" json:"name"`
	ParentID uint64 `gorm:"default:0" json:"parent_id"`
	Type     string `gorm:"size:20;default:branch" json:"type"` // branch(分支机构) / representative(代表机构)
	// Description 机构简介（协会/总会与分支共用，最长 1024 字，与后台「组织机构」表单 maxlength 一致）。
	// 列类型为 text，容量远大于 1024，无需迁移；长度上限由前端表单限制。
	Description string `gorm:"type:text" json:"description"`
	ContactInfo string `gorm:"size:255" json:"contact_info"`
	Sort        int    `gorm:"default:0" json:"sort"`

	Children []*Organization  `gorm:"-" json:"children,omitempty"`
	Levels   []MemberOrgLevel `gorm:"foreignKey:OrgID" json:"levels,omitempty"`
}

func (Organization) TableName() string {
	return "member_organizations"
}
