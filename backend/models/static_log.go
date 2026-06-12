package models

import "gorm.io/gorm"

type StaticLog struct {
	gorm.Model
	Operation string `gorm:"size:50" json:"operation"`
	PageName  string `gorm:"size:100" json:"pageName"`
	Path      string `gorm:"size:255" json:"path"`
	Duration  string `gorm:"size:20" json:"duration"`
	FileSize  string `gorm:"size:20" json:"fileSize"`
	Operator  string `gorm:"size:50" json:"operator"`
	Status    string `gorm:"size:20" json:"status"`
	Message   string `gorm:"type:text" json:"message"`
}
