package models

import (
	"time"

	"gorm.io/gorm"
)

type Tag struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createTime"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	Color     string         `gorm:"size:20;default:'#409eff'" json:"color"`
	Status    int            `gorm:"default:1" json:"status"`
}
