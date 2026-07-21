package models

// App 应用定义，一个应用可被多个租户开通
type App struct {
	BaseModel
	Code        string `gorm:"size:64;uniqueIndex;comment:应用编码" json:"code"`
	Name        string `gorm:"size:128;comment:应用名称" json:"name"`
	Icon        string `gorm:"size:256;comment:图标" json:"icon"`
	Type        string `gorm:"size:32;comment:接入类型 iframe/proxy/micro" json:"type"`
	FrontendURL string `gorm:"size:512;comment:前端入口地址" json:"frontendUrl"`
	BackendURL  string `gorm:"size:512;comment:后端入口地址，用于代理" json:"backendUrl"`
	ApiPrefix   string `gorm:"size:128;comment:API前缀" json:"apiPrefix"`
	Status      int    `gorm:"default:1;comment:状态 1启用 0禁用" json:"status"`
	Sort        int    `gorm:"default:0;comment:排序" json:"sort"`
	Description string `gorm:"size:512;comment:描述" json:"description"`
}

func (App) TableName() string {
	return "base_app"
}

// AppInstance 应用实例，表示某租户开通了某应用
type AppInstance struct {
	BaseModel
	TenantID uint64 `gorm:"index;comment:租户ID" json:"tenantId"`
	AppID    uint64 `gorm:"index;comment:应用ID" json:"appId"`
	Status   int    `gorm:"default:1;comment:状态 1启用 0禁用" json:"status"`
	Config   string `gorm:"type:text;comment:实例配置JSON" json:"config"`
	App      *App   `gorm:"foreignKey:AppID" json:"app,omitempty"`
}

func (AppInstance) TableName() string {
	return "base_app_instance"
}
