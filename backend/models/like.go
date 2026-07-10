package models

import "time"

type LikeAnalytics struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	ArticleID uint      `json:"articleId"`
	LikedAt   time.Time `gorm:"autoCreateTime" json:"likedAt"`
	IP        string    `gorm:"size:50" json:"ip"`
}
