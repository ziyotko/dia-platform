package models

import (
	"time"
)

// BaseModel provides the common primary key and timestamp columns for all models.
// 本项目统一使用物理删除（硬删），不再使用 GORM 软删（gorm.DeletedAt）。
type BaseModel struct {
	ID        uint64    `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
