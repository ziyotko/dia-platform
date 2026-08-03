package models

import "time"

// SignIn record for meeting attendance
type SignIn struct {
	BaseModel
	MeetingID      uint64     `gorm:"index" json:"meetingId"`
	UserID         uint64     `gorm:"index" json:"userId"`
	RegistrationID uint64     `gorm:"index" json:"registrationId"`
	SignInTime     *time.Time `json:"signInTime"`
	SignOutTime    *time.Time `json:"signOutTime"`
	Duration       int        `gorm:"default:0" json:"duration"` // seconds
	Method         string     `gorm:"size:32" json:"method"`     // qrcode/online
	QRCodeToken    string     `gorm:"size:128;uniqueIndex" json:"qrCodeToken"`

	// Relations
	Meeting      *Meeting      `gorm:"foreignKey:MeetingID" json:"meeting,omitempty"`
	User         *User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Registration *Registration `gorm:"foreignKey:RegistrationID" json:"registration,omitempty"`
}

func (SignIn) TableName() string {
	return "conference_sign_ins"
}
