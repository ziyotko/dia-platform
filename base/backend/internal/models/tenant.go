package models

type Tenant struct {
	BaseModel
	Code         string `gorm:"size:64;uniqueIndex;comment:租户编码" json:"code"`
	Name         string `gorm:"size:128;comment:租户名称" json:"name"`
	Status       int    `gorm:"default:1;comment:状态 1启用 0禁用" json:"status"`
	ContactName  string `gorm:"size:64;comment:联系人" json:"contactName"`
	ContactPhone string `gorm:"size:32;comment:联系电话" json:"contactPhone"`
	Description  string `gorm:"size:512;comment:描述" json:"description"`
}

func (Tenant) TableName() string {
	return "base_tenant"
}

// PlatformTenantID 平台级租户 ID。
// 约定：tenant_id == 0 表示平台级资源（平台超管用户、平台预置菜单/字典/机构/模板等）。
const PlatformTenantID uint64 = 0

// IsPlatformTenant 判断是否处于平台级上下文（即平台超级管理员）。
// 全平台统一使用此函数判定超管，勿再散落写 tenantID == 0。
func IsPlatformTenant(tenantID uint64) bool {
	return tenantID == PlatformTenantID
}
