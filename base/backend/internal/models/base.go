package models

import "time"

// BaseModel 所有模型的公共字段。
// 本项目统一使用物理删除（硬删），不再使用 GORM 软删（gorm.DeletedAt）。
type BaseModel struct {
	ID        uint64    `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// NewInvalidBefore 返回「截断到秒」的当前时间，写入 password_changed_at。
// 必须截断：JWT 的 iat 是秒级精度，带小数秒的失效点会把「改密后同一秒内
// 重新登录」新签发的 Token 也判为旧 Token（登录后立刻 401）。
func NewInvalidBefore() time.Time {
	return time.Now().Truncate(time.Second)
}

// IsTokenStale 判断签发时间早于失效点的 Token 是否应作废。
// 两个时间同为秒级精度：恰好等于失效点（同一秒内签发）视为有效。
func IsTokenStale(issuedAt time.Time, invalidBefore *time.Time) bool {
	return invalidBefore != nil && !invalidBefore.IsZero() && issuedAt.Before(*invalidBefore)
}
