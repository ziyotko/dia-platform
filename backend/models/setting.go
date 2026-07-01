package models

import "gorm.io/gorm"

type Setting struct {
	gorm.Model
	SiteName          string `gorm:"size:100" json:"siteName"`
	Logo              string `gorm:"size:500" json:"logo"`
	Icp               string `gorm:"size:200" json:"icp"`
	Copyright         string `gorm:"size:500" json:"copyright"`
	OrgName           string `gorm:"size:200" json:"orgName"`
	OrgCode           string `gorm:"size:100" json:"orgCode"`
	CaptchaEnabled    bool   `gorm:"default:true" json:"captchaEnabled"`
	LockEnabled       bool   `gorm:"default:true" json:"lockEnabled"`
	MaxFailCount      int    `gorm:"default:5" json:"maxFailCount"`
	LockDuration      int    `gorm:"default:30" json:"lockDuration"`
	MinPasswordLength int    `gorm:"default:8" json:"minPasswordLength"`
	TokenExpire       int    `gorm:"default:24" json:"tokenExpire"`
	SmtpHost          string `gorm:"size:200" json:"smtpHost"`
	SmtpPort          string `gorm:"size:10" json:"smtpPort"`
	FromEmail         string `gorm:"size:200" json:"fromEmail"`
	FromName          string `gorm:"size:100" json:"fromName"`
	EmailPassword     string `gorm:"size:255" json:"emailPassword"`
	Ssl               bool   `gorm:"default:true" json:"ssl"`
	ThemeColor        string `gorm:"size:20;default:#409eff" json:"themeColor"`
	SidebarStyle      string `gorm:"size:20;default:light" json:"sidebarStyle"`
	TagsView          bool   `gorm:"default:true" json:"tagsView"`
	Breadcrumb        bool   `gorm:"default:true" json:"breadcrumb"`
}
