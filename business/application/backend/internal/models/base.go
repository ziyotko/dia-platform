package models

import "time"

// BaseModel 所有模型的公共字段。
// 本项目统一使用物理删除（硬删），不再使用 GORM 软删（gorm.DeletedAt）。
type BaseModel struct {
	ID        uint64    `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Admin roles
const (
	RoleSuperAdmin = "super_admin"
	RoleManager    = "manager"
	RoleReviewer   = "reviewer"
)

// Batch statuses
const (
	BatchStatusDraft     = "draft"
	BatchStatusOpen      = "open"
	BatchStatusReviewing = "reviewing"
	BatchStatusClosed    = "closed"
)

// Application statuses
const (
	AppStatusDraft               = "draft"
	AppStatusSubmitted           = "submitted"
	AppStatusPreliminaryRejected = "preliminary_rejected"
	AppStatusUnderReview         = "under_review"
	AppStatusReviewed            = "reviewed"
	AppStatusPassed              = "passed"
	AppStatusRejected            = "rejected"
	AppStatusPublished           = "published"
	AppStatusCertified           = "certified"
)

// Review assignment statuses
const (
	ReviewStatusPending = "pending"
	ReviewStatusScored  = "scored"
)

// Certificate statuses
const (
	CertStatusDraft  = "draft"
	CertStatusIssued = "issued"
	CertStatusVoid   = "void"
)

// Announcement statuses
const (
	AnnouncementStatusDraft     = "draft"
	AnnouncementStatusPublished = "published"
)
