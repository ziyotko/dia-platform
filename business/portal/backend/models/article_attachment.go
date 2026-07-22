package models

import (
	"time"
)

type ArticleAttachment struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	ArticleID uint      `gorm:"not null;index" json:"articleId"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	URL       string    `gorm:"size:500;not null" json:"url"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"createTime"`
}
