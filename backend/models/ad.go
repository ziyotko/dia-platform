package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Ad struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createTime"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Name      string         `gorm:"size:200;not null" json:"name"`
	PageID    uint           `gorm:"not null;index" json:"pageId"`
	ColumnID  uint           `gorm:"default:0;index" json:"columnId"`
	Image     string         `gorm:"size:500" json:"image"`
	Link      string         `gorm:"size:500" json:"link"`
	Sort      int            `gorm:"default:0" json:"sort"`
	Status     int            `gorm:"default:1" json:"status"`
	StartTime  *time.Time     `json:"startTime"`
	EndTime    *time.Time     `json:"endTime"`
	Author     string         `gorm:"size:100" json:"author"`
	AuthorCode string         `gorm:"size:100" json:"authorCode"`
}

func (a Ad) MarshalJSON() ([]byte, error) {
	type Alias Ad
	startTimeStr := ""
	endTimeStr := ""
	if a.StartTime != nil {
		startTimeStr = a.StartTime.Format("2006-01-02 15:04:05")
	}
	if a.EndTime != nil {
		endTimeStr = a.EndTime.Format("2006-01-02 15:04:05")
	}
	return json.Marshal(&struct {
		CreatedAt string `json:"createTime"`
		UpdatedAt string `json:"updatedAt"`
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
		Alias
	}{
		CreatedAt: a.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: a.UpdatedAt.Format("2006-01-02 15:04:05"),
		StartTime: startTimeStr,
		EndTime:   endTimeStr,
		Alias:     (Alias)(a),
	})
}
