package models

import "time"

// BaseModel 所有模型的公共字段。
// 本项目统一使用物理删除（硬删），不再使用 GORM 软删（gorm.DeletedAt）。
type BaseModel struct {
	ID        uint64    `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
