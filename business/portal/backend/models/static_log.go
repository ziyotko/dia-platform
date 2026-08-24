package models

import (
	"time"

	"gorm.io/gorm"
)

type StaticLog struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Operation string         `gorm:"size:50" json:"operation"`
	PageName  string         `gorm:"size:100" json:"pageName"`
	Path      string         `gorm:"size:255" json:"path"`
	Duration  string         `gorm:"size:20" json:"duration"`
	FileSize  string         `gorm:"size:20" json:"fileSize"`
	Operator  string         `gorm:"size:50" json:"operator"`
	Status    string         `gorm:"size:20" json:"status"`
	JobID     string         `gorm:"size:64;index" json:"jobId"`
	Message   string         `gorm:"type:text" json:"message"`
}
