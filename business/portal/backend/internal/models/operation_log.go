package models

import (
	"time"
)

type OperationLog struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	UserID      uint      `gorm:"index" json:"userId"`
	Username    string    `gorm:"size:50" json:"username"`
	Type        string    `gorm:"size:20" json:"type"`
	Module      string    `gorm:"size:50" json:"module"`
	Description string    `gorm:"size:255" json:"description"`
	Method      string    `gorm:"size:10" json:"method"`
	Path        string    `gorm:"size:255" json:"path"`
	Params      string    `gorm:"type:text" json:"params"`
	IP          string    `gorm:"size:50" json:"ip"`
	UserAgent   string    `gorm:"size:500" json:"ua"`
	Duration    int64     `json:"duration"`
	StatusCode  int       `json:"statusCode"`
}
