package models

import (
	"time"
)

type ArticleColumnPublish struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	TemplateID   uint      `gorm:"not null;index" json:"templateId"`
	ColumnID     uint      `gorm:"not null;index" json:"columnId"`
	Column       Column    `gorm:"foreignKey:ColumnID" json:"column"`
	ArticleID    uint      `gorm:"not null;index" json:"articleId"`
	ArticleTitle string    `gorm:"size:200" json:"articleTitle"`
	Author       string    `gorm:"size:100" json:"author"`
	Source       string    `gorm:"size:200" json:"source"`
	IsTop        int       `gorm:"default:0" json:"isTop"`
	IsBold       int       `gorm:"default:0" json:"isBold"`
	Color        string    `gorm:"size:20" json:"color"`
}
