package models

type Tenant struct {
	BaseModel
	Code        string `gorm:"size:64;uniqueIndex;comment:租户编码" json:"code"`
	Name        string `gorm:"size:128;comment:租户名称" json:"name"`
	Status      int    `gorm:"default:1;comment:状态 1启用 0禁用" json:"status"`
	ContactName string `gorm:"size:64;comment:联系人" json:"contactName"`
	ContactPhone string `gorm:"size:32;comment:联系电话" json:"contactPhone"`
	Description string `gorm:"size:512;comment:描述" json:"description"`
}

func (Tenant) TableName() string {
	return "base_tenant"
}
