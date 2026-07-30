package models

import "time"

// MemberOrganization represents member-organization relationship
type MemberOrganization struct {
	BaseModel
	MemberID uint64       `gorm:"index;not null" json:"member_id"`
	OrgID    uint64       `gorm:"index;not null" json:"org_id"`
	JoinedAt time.Time    `gorm:"autoCreateTime" json:"joined_at"`
	Org      Organization `gorm:"foreignKey:OrgID" json:"org,omitempty"`
}

func (MemberOrganization) TableName() string {
	return "member_user_orgs"
}
