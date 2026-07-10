package models

import "time"

type VisitAnalytics struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	ArticleID uint      `json:"articleId"`
	VisitedAt time.Time `gorm:"autoCreateTime" json:"visitedAt"`
	IP        string    `gorm:"size:50" json:"ip"`
}
