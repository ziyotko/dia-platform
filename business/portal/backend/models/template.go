package models

import (
	"gorm.io/gorm"
)

type Template struct {
	gorm.Model
	Name        string `gorm:"size:100;not null" json:"name"`
	Type        string `gorm:"size:20;not null" json:"type"`
	Description string `gorm:"size:500" json:"description"`
	Status      int    `gorm:"default:1" json:"status"`
	SourceCode  string `gorm:"type:text" json:"sourceCode"`
	Layout      string `gorm:"type:text" json:"layout"`
}
