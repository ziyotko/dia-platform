package models

import (
	"gorm.io/gorm"
)

type Department struct {
	gorm.Model
	ParentID    uint   `gorm:"default:0;index" json:"parentId"`
	OrgID       uint   `gorm:"default:0;index:idx_department_org_status" json:"orgId"`
	Name        string `gorm:"size:50;not null" json:"name"`
	Code        string `gorm:"unique;size:50;not null" json:"code"`
	Leader      string `gorm:"size:50" json:"leader"`
	LeaderCode  string `gorm:"size:50" json:"leaderCode"`
	Sort        int    `gorm:"default:0" json:"sort"`
	Status      int    `gorm:"default:1;index:idx_department_org_status" json:"status"`
	Description string `gorm:"size:255" json:"description"`
	UserCount   int    `gorm:"default:0" json:"userCount"`
	UserIds     string `gorm:"size:500" json:"userIds"`
}
