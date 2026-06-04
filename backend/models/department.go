package models

import (
	"time"

	"gorm.io/gorm"
)

type Department struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	ParentID    uint           `gorm:"default:0;index" json:"parentId"`
	Name        string         `gorm:"size:50;not null" json:"name"`
	Code        string         `gorm:"unique;size:50;not null" json:"code"`
	Leader      string         `gorm:"size:50" json:"leader"`
	LeaderCode  string         `gorm:"size:50" json:"leaderCode"`
	Sort        int            `gorm:"default:0" json:"sort"`
	Status      int            `gorm:"default:1" json:"status"`
	Description string         `gorm:"size:255" json:"description"`
	UserCount   int            `gorm:"default:0" json:"userCount"`
	UserIds     string         `gorm:"size:500" json:"userIds"`
}
