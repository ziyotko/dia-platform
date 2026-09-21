package models

import (
	"strings"
	"time"

	"server/utils"

	"gorm.io/gorm"
)

type User struct {
	ID             uint       `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	Username       string     `gorm:"size:50" json:"username"`
	Email          string     `gorm:"unique;not null;size:100" json:"email"`
	Password       string     `gorm:"not null;size:255" json:"-"`
	Account        string     `gorm:"unique;size:50" json:"account"`
	Mobile         string     `gorm:"unique;size:20" json:"mobile"`
	Sex            int        `gorm:"default:0" json:"sex"`
	Status         int        `gorm:"default:1;index" json:"status"`
	RoleIds        string     `gorm:"size:255" json:"roleIds"`
	Bio            string     `gorm:"size:500" json:"bio"`
	Avatar         string     `gorm:"size:500" json:"avatar"`
	LoginFailCount int        `gorm:"default:0" json:"loginFailCount"`
	LockedUntil    *time.Time `json:"lockedUntil"`
	// PasswordChangedAt 最后一次修改密码的时间（本人改密或管理员重置时写入）。
	// 用途：JWT 没有版本号，改密本身不会使已签发的 Token 失效；
	// AuthMiddleware 会比较 Token 的 iat 与该项，使改密前签发的 Token 立即作废。
	PasswordChangedAt *time.Time `json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.Password != "" {
		u.Password = utils.SM3HashPassword(u.Password)
	}
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	// 「是否已哈希」的判据是「含冒号」：SM3HashPassword 的产物形如 "salt:hash"（97 字符）。
	// 原判据 `len(u.Password) != 64` 恒为真，一旦出现带 password 的 Save(&user)/Updates(struct) 调用，
	// 会把已哈希值再次哈希，导致该账号永久无法登录。
	if u.Password != "" && !strings.Contains(u.Password, ":") {
		u.Password = utils.SM3HashPassword(u.Password)
		// 通过模型保存改密时同样记录改密时间，保证「改密后旧 Token 立即失效」的语义
		// 不依赖调用方是否显式写入 password_changed_at。
		now := time.Now().Truncate(time.Second)
		u.PasswordChangedAt = &now
	}
	return nil
}

func (u *User) ComparePassword(password string) bool {
	return utils.VerifySM3Password(password, u.Password)
}
