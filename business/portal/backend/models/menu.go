package models

import (
	"time"

	"gorm.io/gorm"
)

type Menu struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	ParentID  uint           `gorm:"index;default:0" json:"parentId"`
	Name      string         `gorm:"size:50;not null" json:"name"`
	Path      string         `gorm:"size:100" json:"path"`
	Component string         `gorm:"size:200" json:"component"`
	APIPrefix string         `gorm:"size:100" json:"apiPrefix"`
	Icon      string         `gorm:"size:50" json:"icon"`
	Type      string         `gorm:"size:20;default:'directory'" json:"type"`
	Sort      int            `gorm:"default:0" json:"sort"`
	Status    int            `gorm:"default:1" json:"status"`
	Children  []Menu         `gorm:"-" json:"children"`
}
