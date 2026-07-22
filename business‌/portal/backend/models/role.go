package models

import (
	"time"

	"gorm.io/gorm"
)

type Role struct {
	gorm.Model
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Name        string         `gorm:"size:50;not null" json:"name"`
	Code        string         `gorm:"unique;size:50;not null" json:"code"`
	Description string         `gorm:"size:255" json:"description"`
	Status      int            `gorm:"default:1" json:"status"`
	Permissions string         `gorm:"size:500" json:"permissions"`
}
