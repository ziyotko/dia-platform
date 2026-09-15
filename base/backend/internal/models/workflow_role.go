package models

// WorkflowRole 流程角色。
//
// 与系统角色（Role）刻意解耦：Role 决定「能看哪些菜单、能调哪些接口」，
// 流程角色只描述「谁有资格审批某类业务」（如审核组、复审组），
// 供后续审批流程节点选择审批人使用，与菜单/权限无关。
type WorkflowRole struct {
	BaseModel
	TenantID    uint64 `gorm:"index;comment:租户ID" json:"tenantId"`
	Code        string `gorm:"size:64;comment:角色编码" json:"code"`
	Name        string `gorm:"size:128;comment:角色名称" json:"name"`
	Description string `gorm:"size:512;comment:角色说明" json:"description"`
	Status      int    `gorm:"default:1;comment:状态 1启用 0禁用" json:"status"`

	// Users 角色成员（多对多，关联表 base_workflow_role_user）
	Users []User `gorm:"many2many:base_workflow_role_user;" json:"users,omitempty"`

	// UserCount 成员数量，仅列表展示用，不落库（避免每次统计都 join）
	UserCount int64 `gorm:"-" json:"userCount"`
}

func (WorkflowRole) TableName() string {
	return "base_workflow_role"
}

// WorkflowRoleUser 流程角色与用户的关联表。
// 显式声明保证 AutoMigrate 生成的表名/列名与 many2many 关联一致。
type WorkflowRoleUser struct {
	WorkflowRoleID uint64 `gorm:"primaryKey;comment:流程角色ID"`
	UserID         uint64 `gorm:"primaryKey;comment:用户ID"`
}

func (WorkflowRoleUser) TableName() string {
	return "base_workflow_role_user"
}
