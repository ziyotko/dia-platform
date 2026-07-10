package models

import "time"

type ShareAnalytics struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	ArticleID uint      `json:"articleId"`
	SharedAt  time.Time `gorm:"autoCreateTime" json:"sharedAt"`
	IP        string    `gorm:"size:50" json:"ip"`
}
