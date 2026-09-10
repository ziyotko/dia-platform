package models

import (
	"time"
)

// Announcement is a public result announcement (结果公示)
type Announcement struct {
	BaseModel
	BatchID     uint64     `gorm:"index" json:"batchId"`
	Title       string     `gorm:"size:256" json:"title"`
	Content     string     `gorm:"type:text" json:"content"`
	Status      string     `gorm:"size:32;default:draft" json:"status"`
	PublishedAt *time.Time `json:"publishedAt"`

	Batch *ProjectBatch `gorm:"foreignKey:BatchID" json:"batch,omitempty"`
}

func (Announcement) TableName() string {
	return "application_announcements"
}
