package models

import (
	"time"
)

// Notification is a site message sent to an applicant (user_id=0 is reserved and
// never inserted; broadcasts create one row per applicant).
type Notification struct {
	BaseModel
	UserID  uint64     `gorm:"index" json:"userId"`
	Title   string     `gorm:"size:256" json:"title"`
	Content string     `gorm:"type:text" json:"content"`
	Type    string     `gorm:"size:32" json:"type"`
	ReadAt  *time.Time `json:"readAt"`
}

func (Notification) TableName() string {
	return "application_notifications"
}
