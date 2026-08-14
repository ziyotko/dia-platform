package models

import (
	"gorm.io/gorm"
)

type Role struct {
	gorm.Model
	Name        string `gorm:"size:50;not null" json:"name"`
	Code        string `gorm:"unique;size:50;not null" json:"code"`
	Description string `gorm:"size:255" json:"description"`
	Status      int    `gorm:"default:1" json:"status"`
	Permissions string `gorm:"size:500" json:"permissions"`
}
