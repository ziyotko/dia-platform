package models

import "time"

type LoginLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `gorm:"index" json:"createdAt"`
	Username  string    `gorm:"size:100" json:"username"`
	// UserID 登录成功时对应的用户主键（失败/历史行为 0）。
	// 用途：账号显示名（username）并不唯一，仅按用户名过滤会把同名账号的登录记录误发给非管理员；
	// 仪表盘「我的登录日志」因此同时限定 user_id。
	UserID  uint   `gorm:"index" json:"-"`
	IP      string `gorm:"size:50" json:"ip"`
	Browser string `gorm:"size:100" json:"browser"`
	OS      string `gorm:"size:100" json:"os"`
	Device  string `gorm:"size:100" json:"device"`
	Status  int    `gorm:"default:0" json:"status"` // 0失败 1成功
}
