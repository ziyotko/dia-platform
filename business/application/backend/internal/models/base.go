package models

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint64         `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
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
	BatchStatusPublished = "published"
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
)

// Announcement statuses
const (
	AnnouncementStatusDraft     = "draft"
	AnnouncementStatusPublished = "published"
)
