package models

// ApplicationStatus constants
const (
	AppStatusDraft         = "draft"          // 草稿
	AppStatusPendingReview = "pending_review" // 待审核
	AppStatusApproved      = "approved"       // 已通过
	AppStatusRejected      = "rejected"       // 已拒绝
)

// Application represents a membership application
type Application struct {
	BaseModel
	MemberID      uint64       `gorm:"index;not null" json:"member_id"`
	Member        Member       `gorm:"foreignKey:MemberID" json:"member,omitempty"`
	OrgID         uint64       `gorm:"index" json:"org_id"`
	Org           Organization `gorm:"foreignKey:OrgID" json:"org,omitempty"`
	Status        string       `gorm:"size:20;default:draft" json:"status"`
	FormData      string       `gorm:"type:text" json:"form_data"`
	SignedFile    string       `gorm:"size:255" json:"signed_file"`
	ReviewComment string       `gorm:"size:500" json:"review_comment"`
	ReviewerID    uint64       `gorm:"default:0" json:"reviewer_id"`
}

func (Application) TableName() string {
	return "member_applications"
}
