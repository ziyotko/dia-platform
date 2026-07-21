package models

type Organization struct {
	BaseModel
	TenantID    uint64 `gorm:"index;comment:租户ID" json:"tenantId"`
	ParentID    uint64 `gorm:"default:0;comment:父机构ID" json:"parentId"`
	Code        string `gorm:"size:64;comment:机构编码" json:"code"`
	Name        string `gorm:"size:128;comment:机构名称" json:"name"`
	Leader      string `gorm:"size:64;comment:负责人" json:"leader"`
	Phone       string `gorm:"size:32;comment:联系电话" json:"phone"`
	Email       string `gorm:"size:128;comment:邮箱" json:"email"`
	Sort        int    `gorm:"default:0;comment:排序" json:"sort"`
	Status      int    `gorm:"default:1;comment:状态" json:"status"`
	Description string `gorm:"size:512;comment:描述" json:"description"`
	Children    []Organization `gorm:"-" json:"children,omitempty"`
}

func (Organization) TableName() string {
	return "base_organization"
}
