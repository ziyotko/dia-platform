package models

import (
	"time"

	"server/utils"

	"gorm.io/gorm"
)

type User struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time      `json:"createTime"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Username       string         `gorm:"size:50" json:"username"`
	Nickname       string         `gorm:"size:50" json:"nickname"`
	Email          string         `gorm:"unique;not null;size:100" json:"email"`
	Password       string         `gorm:"not null;size:255" json:"-"`
	Account        string         `gorm:"unique;size:50" json:"account"`
	Mobile         string         `gorm:"unique;size:20" json:"mobile"`
	Sex            int            `gorm:"default:0" json:"sex"`
	Status         int            `gorm:"default:1" json:"status"`
	RoleIds        string         `gorm:"size:255" json:"roleIds"`
	Bio            string         `gorm:"size:500" json:"bio"`
	Avatar         string         `gorm:"size:500" json:"avatar"`
	LoginFailCount int            `gorm:"default:0" json:"loginFailCount"`
	LockedUntil    *time.Time     `json:"lockedUntil"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.Password != "" {
		u.Password = utils.SM3HashPassword(u.Password)
	}
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	if u.Password != "" && len(u.Password) != 64 {
		u.Password = utils.SM3HashPassword(u.Password)
	}
	return nil
}

func (u *User) ComparePassword(password string) bool {
	return utils.VerifySM3Password(password, u.Password)
}
