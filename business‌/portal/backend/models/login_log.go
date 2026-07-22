package models

import "time"

type LoginLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `gorm:"index" json:"createTime"`
	Username  string    `gorm:"size:100" json:"username"`
	IP        string    `gorm:"size:50" json:"ip"`
	Browser   string    `gorm:"size:100" json:"browser"`
	OS        string    `gorm:"size:100" json:"os"`
	Device    string    `gorm:"size:100" json:"device"`
	Status    int       `gorm:"default:0" json:"status"` // 0失败 1成功
}
