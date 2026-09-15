package models

type Role struct {
	BaseModel
	TenantID uint64       `gorm:"index;comment:租户ID" json:"tenantId"`
	Code     string       `gorm:"size:64;comment:角色编码" json:"code"`
	Name     string       `gorm:"size:128;comment:角色名称" json:"name"`
	Status   int          `gorm:"default:1;comment:状态" json:"status"`
	Remark   string       `gorm:"size:512;comment:备注" json:"remark"`
	Menus    []Menu       `gorm:"many2many:base_role_menu;" json:"menus,omitempty"`
	Perms    []Permission `gorm:"many2many:base_role_permission;" json:"permissions,omitempty"`
}

func (Role) TableName() string {
	return "base_role"
}

type Permission struct {
	BaseModel
	AppCode  string       `gorm:"size:64;index;comment:所属应用编码" json:"appCode"`
	Code     string       `gorm:"size:128;comment:权限编码" json:"code"`
	Name     string       `gorm:"size:128;comment:权限名称" json:"name"`
	Type     string       `gorm:"size:32;comment:类型 menu/api/button" json:"type"`
	ParentID uint64       `gorm:"default:0;comment:父ID" json:"parentId"`
	Path     string       `gorm:"size:256;comment:路径/API" json:"path"`
	Method   string       `gorm:"size:16;comment:HTTP方法" json:"method"`
	Status   int          `gorm:"default:1;comment:状态" json:"status"`
	Children []Permission `gorm:"-" json:"children,omitempty"`
}

func (Permission) TableName() string {
	return "base_permission"
}

type RolePermission struct {
	RoleID       uint64 `gorm:"primaryKey"`
	PermissionID uint64 `gorm:"primaryKey"`
}

// TableName 与 Role.Perms 的 many2many 关联表保持一致，
// 否则 AutoMigrate 会额外建一张无用的 role_permissions 空表。
func (RolePermission) TableName() string {
	return "base_role_permission"
}
