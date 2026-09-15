package models

import "time"

// 流程实例状态
const (
	WorkflowInstanceRunning  = 1 // 审批中
	WorkflowInstanceApproved = 2 // 已通过
	WorkflowInstanceRejected = 3 // 已驳回
	WorkflowInstanceCanceled = 4 // 已撤销
)

// 审批任务状态
const (
	WorkflowTaskPending  = 1 // 待处理
	WorkflowTaskApproved = 2 // 已通过
	WorkflowTaskRejected = 3 // 已驳回
	WorkflowTaskInvalid  = 4 // 已失效（他人已处理 / 流程终止）
)

// 节点审批人类型
const (
	ApproverTypeRole      = "role"      // 流程角色（角色内成员或签）
	ApproverTypeUser      = "user"      // 指定用户
	ApproverTypeInitiator = "initiator" // 发起人本人
)

// 流程日志动作
const (
	WorkflowActionStart    = "start"     // 发起
	WorkflowActionApprove  = "approve"   // 通过
	WorkflowActionReject   = "reject"    // 驳回
	WorkflowActionCancel   = "cancel"    // 撤销
	WorkflowActionAutoPass = "auto-pass" // 节点无有效审批人，自动通过
	WorkflowActionFinish   = "finish"    // 流程结束（通过）
)

// Workflow 流程定义。
// 一个流程由若干按 Sort 顺序执行的审批节点串联而成（串行或签）。
type Workflow struct {
	BaseModel
	TenantID    uint64 `gorm:"index;comment:租户ID" json:"tenantId"`
	Code        string `gorm:"size:64;comment:流程编码" json:"code"`
	Name        string `gorm:"size:128;comment:流程名称" json:"name"`
	Description string `gorm:"size:512;comment:流程说明" json:"description"`
	Status      int    `gorm:"default:1;comment:状态 1启用 0停用" json:"status"`

	// Nodes 节点列表（保存时按 Sort 重排，展示用）
	Nodes []WorkflowNode `gorm:"foreignKey:WorkflowID" json:"nodes,omitempty"`
}

func (Workflow) TableName() string {
	return "base_workflow"
}

// WorkflowNode 流程节点（审批环节）。
// 同一节点内的多个审批人为「或签」：任一审批人处理即视为该节点完成，其余任务自动失效。
type WorkflowNode struct {
	BaseModel
	WorkflowID   uint64 `gorm:"index;comment:流程定义ID" json:"workflowId"`
	Name         string `gorm:"size:128;comment:节点名称" json:"name"`
	Sort         int    `gorm:"default:1;comment:节点顺序" json:"sort"`
	ApproverType string `gorm:"size:32;comment:审批人类型 role/user/initiator" json:"approverType"`
	ApproverID   uint64 `gorm:"default:0;comment:流程角色ID或用户ID" json:"approverId"`
	Description  string `gorm:"size:512;comment:节点说明" json:"description"`

	// ApproverName 审批人展示名（流程角色名/用户名/发起人），不落库
	ApproverName string `gorm:"-" json:"approverName,omitempty"`
}

func (WorkflowNode) TableName() string {
	return "base_workflow_node"
}

// WorkflowInstance 流程实例。
// business_type / business_id 用于关联底座或子应用的业务单据（可为空，表示纯手工发起）。
type WorkflowInstance struct {
	BaseModel
	TenantID      uint64    `gorm:"index;comment:租户ID" json:"tenantId"`
	WorkflowID    uint64    `gorm:"index;comment:流程定义ID" json:"workflowId"`
	WorkflowName  string    `gorm:"size:128;comment:流程名称快照" json:"workflowName"`
	Title         string    `gorm:"size:256;comment:标题" json:"title"`
	BusinessType  string    `gorm:"size:64;comment:业务类型" json:"businessType"`
	BusinessID    string    `gorm:"size:64;comment:业务ID" json:"businessId"`
	Content       string    `gorm:"type:text;comment:申请说明" json:"content"`
	InitiatorID   uint64    `gorm:"index;comment:发起人ID" json:"initiatorId"`
	InitiatorName string    `gorm:"size:64;comment:发起人" json:"initiatorName"`
	CurrentSort   int       `gorm:"default:0;comment:当前节点顺序" json:"currentSort"`
	CurrentNode   string    `gorm:"size:128;comment:当前节点名称" json:"currentNode"`
	Status        int       `gorm:"default:1;comment:状态 1审批中 2已通过 3已驳回 4已撤销" json:"status"`
	StartAt       time.Time `gorm:"comment:发起时间" json:"startAt"`
	// EndAt 为指针类型：未结束时写入 NULL，避免 GORM 写入 0000-00-00 被 MySQL 严格模式拒绝
	EndAt *time.Time `gorm:"comment:结束时间" json:"endAt"`
}

func (WorkflowInstance) TableName() string {
	return "base_workflow_instance"
}

// WorkflowTask 审批任务（待办）。
type WorkflowTask struct {
	BaseModel
	TenantID     uint64 `gorm:"index;comment:租户ID" json:"tenantId"`
	InstanceID   uint64 `gorm:"index;comment:流程实例ID" json:"instanceId"`
	NodeID       uint64 `gorm:"index;comment:节点ID" json:"nodeId"`
	NodeName     string `gorm:"size:128;comment:节点名称快照" json:"nodeName"`
	NodeSort     int    `gorm:"comment:节点顺序快照" json:"nodeSort"`
	ApproverID   uint64 `gorm:"index;comment:审批人ID" json:"approverId"`
	ApproverName string `gorm:"size:64;comment:审批人" json:"approverName"`
	Status       int    `gorm:"default:1;comment:状态 1待处理 2已通过 3已驳回 4已失效" json:"status"`
	Comment      string `gorm:"size:512;comment:审批意见" json:"comment"`
	// HandledAt 为指针类型：未处理时写入 NULL（同 EndAt，避免零值时间被 MySQL 拒绝）
	HandledAt *time.Time `gorm:"comment:处理时间" json:"handledAt"`

	// Instance 待办列表展示所需的流程实例信息（belongs-to，查询时 Preload）
	Instance *WorkflowInstance `gorm:"foreignKey:InstanceID" json:"instance,omitempty"`
}

func (WorkflowTask) TableName() string {
	return "base_workflow_task"
}

// WorkflowLog 流程流转日志（审批时间线）。
type WorkflowLog struct {
	BaseModel
	TenantID     uint64 `gorm:"index;comment:租户ID" json:"tenantId"`
	InstanceID   uint64 `gorm:"index;comment:流程实例ID" json:"instanceId"`
	NodeID       uint64 `gorm:"comment:节点ID" json:"nodeId"`
	NodeName     string `gorm:"size:128;comment:节点名称" json:"nodeName"`
	OperatorID   uint64 `gorm:"comment:操作人ID" json:"operatorId"`
	OperatorName string `gorm:"size:64;comment:操作人" json:"operatorName"`
	Action       string `gorm:"size:32;comment:动作 start/approve/reject/cancel/auto-pass/finish" json:"action"`
	Comment      string `gorm:"size:512;comment:意见/说明" json:"comment"`
}

func (WorkflowLog) TableName() string {
	return "base_workflow_log"
}
