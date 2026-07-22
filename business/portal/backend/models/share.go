package models

import "time"

type ShareAnalytics struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	ArticleID uint      `gorm:"index" json:"articleId"`
	SharedAt  time.Time `gorm:"autoCreateTime;index" json:"sharedAt"`
	IP        string    `gorm:"size:50" json:"ip"`
}
