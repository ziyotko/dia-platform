package models

type Menu struct {
	BaseModel
	TenantID  uint64  `gorm:"index;comment:租户ID，0表示系统默认" json:"tenantId"`
	AppCode   string  `gorm:"size:64;index;comment:所属应用编码，base表示底座" json:"appCode"`
	ParentID  uint64  `gorm:"default:0;comment:父菜单ID" json:"parentId"`
	Name      string  `gorm:"size:128;comment:菜单名称" json:"name"`
	Icon      string  `gorm:"size:64;comment:图标" json:"icon"`
	Path      string  `gorm:"size:256;comment:路由路径" json:"path"`
	Component string  `gorm:"size:256;comment:组件路径" json:"component"`
	Type      string  `gorm:"size:32;comment:类型 directory/menu/button" json:"type"`
	Permission string `gorm:"size:128;comment:权限标识" json:"permission"`
	Sort      int     `gorm:"default:0;comment:排序" json:"sort"`
	Status    int     `gorm:"default:1;comment:状态" json:"status"`
	Hidden    bool    `gorm:"default:false;comment:是否隐藏" json:"hidden"`
	KeepAlive bool    `gorm:"default:false;comment:是否缓存" json:"keepAlive"`
	Target    string  `gorm:"size:32;comment:打开方式 _self/_blank/iframe" json:"target"`
	Children  []Menu  `gorm:"-" json:"children,omitempty"`
}

func (Menu) TableName() string {
	return "base_menu"
}
