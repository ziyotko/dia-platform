package models

import (
	"time"
)

// ProjectBatch is an application round (申报批次/申报通知)
type ProjectBatch struct {
	BaseModel
	Title          string     `gorm:"size:256" json:"title"`
	CategoryID     uint64     `gorm:"index" json:"categoryId"`
	Description    string     `gorm:"type:text" json:"description"`
	Requirements   string     `gorm:"type:text" json:"requirements"`
	ApplyStart     *time.Time `json:"applyStart"`
	ApplyEnd       *time.Time `json:"applyEnd"`
	ReviewDeadline *time.Time `json:"reviewDeadline"`
	Status         string     `gorm:"size:32;default:draft" json:"status"`

	Category *ProjectCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`

	// Transient, filled by the service layer: whether the applicant may still
	// submit into this batch right now.
	CanApply bool `gorm:"-" json:"canApply"`
}

func (ProjectBatch) TableName() string {
	return "application_project_batches"
}
