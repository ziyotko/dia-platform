package models

import (
	"time"
)

// ReviewAssignment assigns a reviewer (评审人) to an application
type ReviewAssignment struct {
	BaseModel
	ApplicationID uint64     `gorm:"index" json:"applicationId"`
	ReviewerID    uint64     `gorm:"index" json:"reviewerId"`
	Status        string     `gorm:"size:32;default:pending" json:"status"`
	Score         float64    `gorm:"type:decimal(6,2);default:0" json:"score"`
	Comment       string     `gorm:"type:text" json:"comment"`
	ReviewedAt    *time.Time `json:"reviewedAt"`

	Reviewer *Admin `gorm:"foreignKey:ReviewerID" json:"reviewer,omitempty"`
}

func (ReviewAssignment) TableName() string {
	return "application_review_assignments"
}
