package models

import "time"

// Notification message
type Notification struct {
	BaseModel
	SenderID   uint64     `gorm:"index" json:"senderId"`
	Title      string     `gorm:"size:256" json:"title"`
	Content    string     `gorm:"type:text" json:"content"`
	TargetType string     `gorm:"size:32" json:"targetType"`  // all/registered/branch/level/specific
	TargetIDs  string     `gorm:"type:text" json:"targetIds"` // JSON array
	SentAt     *time.Time `json:"sentAt"`
}

func (Notification) TableName() string {
	return "conference_notifications"
}

// NotificationRead tracks read status
type NotificationRead struct {
	BaseModel
	NotificationID uint64     `gorm:"index" json:"notificationId"`
	UserID         uint64     `gorm:"index" json:"userId"`
	ReadAt         *time.Time `json:"readAt"`

	Notification *Notification `gorm:"foreignKey:NotificationID" json:"notification,omitempty"`
}

func (NotificationRead) TableName() string {
	return "conference_notification_reads"
}
