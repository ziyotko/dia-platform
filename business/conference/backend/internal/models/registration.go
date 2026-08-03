package models

import "time"

// Registration for a meeting
type Registration struct {
	BaseModel
	MeetingID        uint64     `gorm:"index" json:"meetingId"`
	UserID           uint64     `gorm:"index" json:"userId"`
	Status           string     `gorm:"size:32;default:pending" json:"status"` // pending/approved/rejected/waitlist/cancelled
	WaitlistPosition int        `gorm:"default:0" json:"waitlistPosition"`
	ReviewComment    string     `gorm:"size:512" json:"reviewComment"`
	ReviewedBy       uint64     `gorm:"default:0" json:"reviewedBy"`
	ReviewedAt       *time.Time `json:"reviewedAt"`

	// Relations
	Meeting *Meeting `gorm:"foreignKey:MeetingID" json:"meeting,omitempty"`
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Registration) TableName() string {
	return "conference_registrations"
}
