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

// Meeting types
const (
	MeetingTypeOnline  = "online"
	MeetingTypeOffline = "offline"
	MeetingTypeHybrid  = "hybrid"
)

// Meeting statuses
const (
	MeetingStatusDraft    = "draft"
	MeetingStatusOpen     = "open"
	MeetingStatusClosed   = "closed"
	MeetingStatusArchived = "archived"
)

// Registration statuses
const (
	RegStatusPending   = "pending"
	RegStatusApproved  = "approved"
	RegStatusRejected  = "rejected"
	RegStatusWaitlist  = "waitlist"
	RegStatusCancelled = "cancelled"
)

// Payment statuses
const (
	PayStatusUnpaid    = "unpaid"
	PayStatusPaid      = "paid"
	PayStatusRefunding = "refunding"
	PayStatusRefunded  = "refunded"
)

// Refund statuses
const (
	RefundStatusPending  = "pending"
	RefundStatusApproved = "approved"
	RefundStatusRejected = "rejected"
)

// Vote types
const (
	VoteTypeSingle       = "single"
	VoteTypeMultiple     = "multiple"
	VoteTypeEqual        = "equal"
	VoteTypeDifferential = "differential"
)

// Survey question types
const (
	SurveyQTypeSingle = "single"
	SurveyQTypeMulti  = "multi"
	SurveyQTypeText   = "text"
)

// Admin roles
const (
	RoleSuperAdmin   = "super_admin"
	RoleMeetingAdmin = "meeting_admin"
	RoleFinance      = "finance"
	RoleBranchAdmin  = "branch_admin"
	RoleSupervisor   = "supervisor"
)

// Invoice types
const (
	InvoiceTypeIndividual = "individual"
	InvoiceTypeCompany    = "company"
)

// Notification target types
const (
	NotifyTargetAll        = "all"
	NotifyTargetRegistered = "registered"
	NotifyTargetBranch     = "branch"
	NotifyTargetLevel      = "level"
	NotifyTargetSpecific   = "specific"
)

// Credit sources
const (
	CreditSourceAuto   = "auto_assign"
	CreditSourceManual = "manual"
)

// Archive item types
const (
	ArchiveItemRegistration = "registration"
	ArchiveItemSignIn       = "signin"
	ArchiveItemFinance      = "finance"
	ArchiveItemVote         = "vote"
	ArchiveItemSurvey       = "survey"
	ArchiveItemMaterial     = "material"
	ArchiveItemMinutes      = "minutes"
	ArchiveItemLive         = "live"
)

// Live statuses
const (
	LiveStatusOff   = "off"
	LiveStatusLive  = "live"
	LiveStatusEnded = "ended"
)
