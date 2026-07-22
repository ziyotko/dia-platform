package models

import (
	"time"

	"gorm.io/gorm"
)

// Organization 机构（支持集团型复杂组织架构）
type Organization struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	ParentID    uint           `gorm:"default:0;index" json:"parentId"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Code        string         `gorm:"unique;size:50;not null" json:"code"`
	OrgType     int            `gorm:"default:3;index:idx_organization_status_org_type" json:"orgType"`
	OrgLevel    int            `gorm:"default:1" json:"orgLevel"`
	Category    string         `gorm:"size:50" json:"category"`
	Region      string         `gorm:"size:100" json:"region"`
	Province    string         `gorm:"size:50" json:"province"`
	City        string         `gorm:"size:50" json:"city"`
	Address     string         `gorm:"size:255" json:"address"`
	Manager     string         `gorm:"size:50" json:"manager"`
	ManagerCode string         `gorm:"size:50" json:"managerCode"`
	Sort        int            `gorm:"default:0" json:"sort"`
	Status      int            `gorm:"default:1;index:idx_organization_status_org_type" json:"status"`
	Description string         `gorm:"size:500" json:"description"`
	UserCount   int            `gorm:"default:0" json:"userCount"`
	UserIds     string         `gorm:"size:1000" json:"userIds"`
}

// OrgTypeText 返回机构类型文本
func (o Organization) OrgTypeText() string {
	switch o.OrgType {
	case 1:
		return "机构"
	case 2:
		return "分支机构"
	default:
		return "其他"
	}
}
