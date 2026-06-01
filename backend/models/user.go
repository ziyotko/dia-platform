package models

import (
	"server/utils"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email    string `gorm:"unique;not null;size:100" json:"email"`
	Password string `gorm:"not null;size:255" json:"-"`
	Username string `gorm:"size:50" json:"username"`
	Account  string `gorm:"unique;size:50" json:"account"`
	Sex      int    `gorm:"default:0" json:"sex"`
	Mobile   string `gorm:"unique;size:20" json:"mobile"`
	Status   int    `gorm:"default:1" json:"status"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.Password != "" {
		u.Password = utils.SM3HashPassword(u.Password)
	}
	return nil
}

func (u *User) ComparePassword(password string) bool {
	return utils.VerifySM3Password(password, u.Password)
}
