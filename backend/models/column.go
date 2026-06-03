package models

import (
	"time"

	"gorm.io/gorm"
)

type Column struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Code        string         `gorm:"size:100;not null" json:"code"`
	PageID      uint           `gorm:"not null;index" json:"pageId"`
	ParentID    uint           `gorm:"default:0;index" json:"parentId"`
	RoutePath   string         `gorm:"size:200" json:"routePath"`
	Template    string         `gorm:"size:100" json:"template"`
	Description string         `gorm:"size:500" json:"description"`
	Sort        int            `gorm:"default:0" json:"sort"`
	Status      int            `gorm:"default:1" json:"status"`
}
