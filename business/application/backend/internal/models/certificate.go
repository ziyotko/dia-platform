package models

import (
	"time"
)

// Certificate is issued to an applicant after their project is approved
type Certificate struct {
	BaseModel
	ApplicationID uint64     `gorm:"index" json:"applicationId"`
	UserID        uint64     `gorm:"index" json:"userId"`
	BatchID       uint64     `gorm:"index" json:"batchId"`
	CertNo        string     `gorm:"size:64" json:"certNo"`
	Title         string     `gorm:"size:256" json:"title"`
	Holder        string     `gorm:"size:128" json:"holder"`
	FileURL       string     `gorm:"size:512" json:"fileUrl"`
	Status        string     `gorm:"size:32;default:draft" json:"status"`
	IssuedAt      *time.Time `json:"issuedAt"`

	Application *Application `gorm:"foreignKey:ApplicationID" json:"application,omitempty"`
}

func (Certificate) TableName() string {
	return "application_certificates"
}
