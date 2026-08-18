package models

import "gorm.io/gorm"

type Setting struct {
	gorm.Model
	SiteName                 string `gorm:"size:100" json:"siteName"`
	Logo                     string `gorm:"size:500" json:"logo"`
	Icp                      string `gorm:"size:200" json:"icp"`
	Copyright                string `gorm:"size:500" json:"copyright"`
	OrgName                  string `gorm:"size:200" json:"orgName"`
	OrgCode                  string `gorm:"size:100" json:"orgCode"`
	CaptchaEnabled           bool   `gorm:"default:true" json:"captchaEnabled"`
	LockEnabled              bool   `gorm:"default:true" json:"lockEnabled"`
	MaxFailCount             int    `gorm:"default:5" json:"maxFailCount"`
	LockDuration             int    `gorm:"default:30" json:"lockDuration"`
	MinPasswordLength        int    `gorm:"default:8" json:"minPasswordLength"`
	TokenExpire              int    `gorm:"default:24" json:"tokenExpire"`
	SmtpHost                 string `gorm:"size:200" json:"smtpHost"`
	SmtpPort                 string `gorm:"size:10" json:"smtpPort"`
	FromEmail                string `gorm:"size:200" json:"fromEmail"`
	FromName                 string `gorm:"size:100" json:"fromName"`
	EmailPassword            string `gorm:"size:255" json:"emailPassword"`
	Ssl                      bool   `gorm:"default:true" json:"ssl"`
	StaticPath               string `gorm:"size:255" json:"staticPath"`
	HomeGray                 bool   `gorm:"default:false" json:"homeGray"`
	HomeStaticTimeEnabled    bool   `gorm:"default:false" json:"homeStaticTimeEnabled"`
	HomeStaticTime           string `gorm:"size:10" json:"homeStaticTime"`
	ColumnStaticTimeEnabled  bool   `gorm:"default:false" json:"columnStaticTimeEnabled"`
	ColumnStaticTime         string `gorm:"size:10" json:"columnStaticTime"`
	SpecialStaticTimeEnabled bool   `gorm:"default:false" json:"specialStaticTimeEnabled"`
	SpecialStaticTime        string `gorm:"size:10" json:"specialStaticTime"`
	DetailStaticTimeEnabled  bool   `gorm:"default:false" json:"detailStaticTimeEnabled"`
	DetailStaticTime         string `gorm:"size:10" json:"detailStaticTime"`
	StaticProgramAddr        string `gorm:"size:100" json:"staticProgramAddr"`
	StaticProgramTokenName   string `gorm:"size:100" json:"staticProgramTokenName"`
}
