package service

import (
	"errors"
	"time"

	"base/internal/models"
	"base/pkg/db"

	"gorm.io/gorm"
)

// WorkflowEngineService 流程引擎：发起、审批、驳回、撤销与查询。
//
// 引擎语义（刻意保持简单、可预测）：
//   - 节点按 sort 顺序串行执行；
//   - 同一节点内多个审批人为**或签**：任一人处理即代表节点完成，其余待办自动失效；
//   - 节点解析不到有效审批人（角色无成员 / 用户被删除或停用）时**自动通过**并记入流转日志，
//     避免流程被永久挂起；
//   - 审批全部节点后实例置为「已通过」；任一节点驳回则实例置为「已驳回」。
type WorkflowEngineService struct{}

// WorkflowActor 当前操作者（来自 JWT 上下文），用于归属校验与日志记录。
type WorkflowActor struct {
	UserID   uint64
	Username string
	TenantID uint64
	IsAdmin  bool
}

// StartRequest 发起流程请求
type StartRequest struct {
	WorkflowID   uint64 `json:"workflowId"`
	Title        string `json:"title"`
	BusinessType string `json:"businessType"`
	BusinessID   string `json:"businessId"`
	Content      string `json:"content"`
}

// WorkflowInstanceQuery 流程实例列表查询条件
type WorkflowInstanceQuery struct {
	TenantID       uint64
	FilterTenantID uint64
	Mine           bool
	UserID         uint64
	WorkflowID     uint64
	Status         *int
	Keyword        string
	Page           int
	Size           int
}

// WorkflowInstanceDetail 实例详情（含任务与流转日志）
type WorkflowInstanceDetail struct {
	Instance models.WorkflowInstance `json:"instance"`
	Tasks    []models.WorkflowTask   `json:"tasks"`
	Logs     []models.WorkflowLog    `json:"logs"`
}

// Start 发起流程：创建实例并推进到第一个有审批人的节点。
func (s WorkflowEngineService) Start(req StartRequest, actor WorkflowActor) (*models.WorkflowInstance, error) {
	if req.Title == "" {
		return nil, errors.New("请填写标题")
	}

	instance := &models.WorkflowInstance{}
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var wf models.Workflow
		query := tx.Preload("Nodes", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort ASC, id ASC")
		}).Where("id = ?", req.WorkflowID)
		// 普通租户只能发起本租户的流程；平台超管可发起任意租户的流程
		if actor.TenantID > 0 {
			query = query.Where("tenant_id = ?", actor.TenantID)
		}
		if err := query.First(&wf).Error; err != nil {
			return errors.New("流程定义不存在")
		}
		if wf.Status != 1 {
			return errors.New("该流程已停用，无法发起")
		}
		if len(wf.Nodes) == 0 {
			return errors.New("该流程未配置审批节点，无法发起")
		}

		name := actorDisplayName(tx, actor)
		instance = &models.WorkflowInstance{
			TenantID:      wf.TenantID,
			WorkflowID:    wf.ID,
			WorkflowName:  wf.Name,
			Title:         req.Title,
			BusinessType:  req.BusinessType,
			BusinessID:    req.BusinessID,
			Content:       req.Content,
			InitiatorID:   actor.UserID,
			InitiatorName: name,
			Status:        models.WorkflowInstanceRunning,
			StartAt:       time.Now(),
		}
		if err := tx.Create(instance).Error; err != nil {
			return err
		}
		if err := createWorkflowLog(tx, instance, 0, "", actor.UserID, name, models.WorkflowActionStart, "发起流程"); err != nil {
			return err
		}
		return advanceWorkflow(tx, instance, 0)
	})
	if err != nil {
		return nil, err
	}
	// 流程已推进到第一个审批节点，重新读取以保证返回体中的 currentSort/currentNode 为最新值
	if err := db.DB.Where("id = ?", instance.ID).First(instance).Error; err != nil {
		return nil, err
	}
	return instance, nil
}

// Approve 审批通过当前任务，并按需推进到下一节点。
func (s WorkflowEngineService) Approve(taskID uint64, comment string, actor WorkflowActor) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		task, instance, err := loadWorkflowTask(tx, taskID)
		if err != nil {
			return err
		}
		if err := ensureApprovable(task, instance, actor); err != nil {
			return err
		}
		name := actorDisplayName(tx, actor)
		now := time.Now()

		if err := tx.Model(&models.WorkflowTask{}).Where("id = ?", task.ID).
			Updates(map[string]interface{}{
				"status":     models.WorkflowTaskApproved,
				"comment":    comment,
				"handled_at": now,
			}).Error; err != nil {
			return err
		}
		// 或签：同一节点其他待办自动失效
		if err := tx.Model(&models.WorkflowTask{}).
			Where("instance_id = ? AND node_id = ? AND id <> ? AND status = ?", instance.ID, task.NodeID, task.ID, models.WorkflowTaskPending).
			Updates(map[string]interface{}{
				"status":     models.WorkflowTaskInvalid,
				"comment":    "同节点其他审批人已处理",
				"handled_at": now,
			}).Error; err != nil {
			return err
		}
		if err := createWorkflowLog(tx, instance, task.NodeID, task.NodeName, actor.UserID, name, models.WorkflowActionApprove, comment); err != nil {
			return err
		}
		return advanceWorkflow(tx, instance, task.NodeSort)
	})
}

// Reject 驳回当前任务：流程立即结束，其余待办失效。
func (s WorkflowEngineService) Reject(taskID uint64, comment string, actor WorkflowActor) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		task, instance, err := loadWorkflowTask(tx, taskID)
		if err != nil {
			return err
		}
		if err := ensureApprovable(task, instance, actor); err != nil {
			return err
		}
		name := actorDisplayName(tx, actor)
		now := time.Now()

		if err := tx.Model(&models.WorkflowTask{}).Where("id = ?", task.ID).
			Updates(map[string]interface{}{
				"status":     models.WorkflowTaskRejected,
				"comment":    comment,
				"handled_at": now,
			}).Error; err != nil {
			return err
		}
		if err := invalidatePendingTasks(tx, instance.ID, task.ID, "流程已驳回", now); err != nil {
			return err
		}
		if err := tx.Model(&models.WorkflowInstance{}).Where("id = ?", instance.ID).
			Updates(map[string]interface{}{
				"status":       models.WorkflowInstanceRejected,
				"current_sort": -1,
				"current_node": "已驳回",
				"end_at":       now,
			}).Error; err != nil {
			return err
		}
		return createWorkflowLog(tx, instance, task.NodeID, task.NodeName, actor.UserID, name, models.WorkflowActionReject, comment)
	})
}

// Cancel 撤销流程：仅发起人本人或管理员可撤销审批中的实例。
func (s WorkflowEngineService) Cancel(instanceID uint64, actor WorkflowActor) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var instance models.WorkflowInstance
		query := tx.Where("id = ?", instanceID)
		if !models.IsPlatformTenant(actor.TenantID) {
			query = query.Where("tenant_id = ?", actor.TenantID)
		}
		if err := query.First(&instance).Error; err != nil {
			return errors.New("流程实例不存在")
		}
		if instance.Status != models.WorkflowInstanceRunning {
			return errors.New("该流程已结束，无法撤销")
		}
		if instance.InitiatorID != actor.UserID && !actor.IsAdmin && !models.IsPlatformTenant(actor.TenantID) {
			return errors.New("只有发起人或管理员可以撤销")
		}

		name := actorDisplayName(tx, actor)
		now := time.Now()
		if err := invalidatePendingTasks(tx, instance.ID, 0, "流程已撤销", now); err != nil {
			return err
		}
		if err := tx.Model(&models.WorkflowInstance{}).Where("id = ?", instance.ID).
			Updates(map[string]interface{}{
				"status":       models.WorkflowInstanceCanceled,
				"current_sort": -1,
				"current_node": "已撤销",
				"end_at":       now,
			}).Error; err != nil {
			return err
		}
		return createWorkflowLog(tx, &instance, 0, "", actor.UserID, name, models.WorkflowActionCancel, "撤销流程")
	})
}

// Delete 删除流程实例（仅管理员）：软删除实例及其任务与流转日志。
func (s WorkflowEngineService) Delete(instanceID uint64, actor WorkflowActor) error {
	if !actor.IsAdmin && !models.IsPlatformTenant(actor.TenantID) {
		return errors.New("只有管理员可以删除流程实例")
	}
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var instance models.WorkflowInstance
		query := tx.Where("id = ?", instanceID)
		if !models.IsPlatformTenant(actor.TenantID) {
			query = query.Where("tenant_id = ?", actor.TenantID)
		}
		if err := query.First(&instance).Error; err != nil {
			return errors.New("流程实例不存在")
		}
		if err := tx.Where("instance_id = ?", instance.ID).Delete(&models.WorkflowTask{}).Error; err != nil {
			return err
		}
		if err := tx.Where("instance_id = ?", instance.ID).Delete(&models.WorkflowLog{}).Error; err != nil {
			return err
		}
		return tx.Delete(&instance).Error
	})
}

// GetInstance 实例详情：管理员、发起人、以及参与审批的人可查看。
func (s WorkflowEngineService) GetInstance(id uint64, actor WorkflowActor) (*WorkflowInstanceDetail, error) {
	var instance models.WorkflowInstance
	query := db.DB.Where("id = ?", id)
	if !models.IsPlatformTenant(actor.TenantID) {
		query = query.Where("tenant_id = ?", actor.TenantID)
	}
	if err := query.First(&instance).Error; err != nil {
		return nil, errors.New("流程实例不存在")
	}

	isAdmin := actor.IsAdmin || models.IsPlatformTenant(actor.TenantID)
	if !isAdmin && instance.InitiatorID != actor.UserID {
		var count int64
		if err := db.DB.Model(&models.WorkflowTask{}).
			Where("instance_id = ? AND approver_id = ?", id, actor.UserID).
			Count(&count).Error; err != nil {
			return nil, err
		}
		if count == 0 {
			return nil, errors.New("无权查看该流程实例")
		}
	}

	detail := &WorkflowInstanceDetail{Instance: instance}
	if err := db.DB.Where("instance_id = ?", id).Order("node_sort ASC, id ASC").Find(&detail.Tasks).Error; err != nil {
		return nil, err
	}
	if err := db.DB.Where("instance_id = ?", id).Order("created_at ASC, id ASC").Find(&detail.Logs).Error; err != nil {
		return nil, err
	}
	return detail, nil
}

// ListInstances 实例列表。普通用户（非管理员）只能查看自己发起的实例。
func (s WorkflowEngineService) ListInstances(q WorkflowInstanceQuery) ([]models.WorkflowInstance, int64, error) {
	var list []models.WorkflowInstance
	var total int64
	query := db.DB.Model(&models.WorkflowInstance{})
	if q.TenantID > 0 {
		query = query.Where("tenant_id = ?", q.TenantID)
	} else if q.FilterTenantID > 0 {
		query = query.Where("tenant_id = ?", q.FilterTenantID)
	}
	if q.Mine {
		query = query.Where("initiator_id = ?", q.UserID)
	}
	if q.WorkflowID > 0 {
		query = query.Where("workflow_id = ?", q.WorkflowID)
	}
	if q.Status != nil {
		query = query.Where("status = ?", *q.Status)
	}
	if q.Keyword != "" {
		query = query.Where("title LIKE ? OR workflow_name LIKE ?", "%"+q.Keyword+"%", "%"+q.Keyword+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("created_at DESC").Offset((q.Page - 1) * q.Size).Limit(q.Size).Find(&list).Error
	return list, total, err
}

// MyTasks 我的审批任务。
// box=todo 待我处理；box=done 我已处理（仅已通过/已驳回，
// 被他人处理而失效的任务不算「我已办」，否则会把从未处理过的任务列进来）。
func (s WorkflowEngineService) MyTasks(actor WorkflowActor, box string, page, size int) ([]models.WorkflowTask, int64, error) {
	var list []models.WorkflowTask
	var total int64
	query := db.DB.Model(&models.WorkflowTask{}).Where("approver_id = ?", actor.UserID)
	if !models.IsPlatformTenant(actor.TenantID) {
		query = query.Where("tenant_id = ?", actor.TenantID)
	}
	if box == "done" {
		query = query.Where("status IN ?", []int{models.WorkflowTaskApproved, models.WorkflowTaskRejected})
	} else {
		query = query.Where("status = ?", models.WorkflowTaskPending)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Preload("Instance").Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

// advanceWorkflow 从 afterSort 之后推进流程（在事务内调用）。
// 跳过没有有效审批人的节点并记入日志，遇到第一个有审批人的节点则创建待办并停下；
// 若已无后续节点，则把实例置为「已通过」。
func advanceWorkflow(tx *gorm.DB, instance *models.WorkflowInstance, afterSort int) error {
	var nodes []models.WorkflowNode
	if err := tx.Where("workflow_id = ? AND sort > ?", instance.WorkflowID, afterSort).
		Order("sort ASC, id ASC").Find(&nodes).Error; err != nil {
		return err
	}

	for _, node := range nodes {
		approvers, err := resolveApprovers(tx, &node, instance.InitiatorID)
		if err != nil {
			return err
		}
		if len(approvers) == 0 {
			if err := createWorkflowLog(tx, instance, node.ID, node.Name, 0, "系统",
				models.WorkflowActionAutoPass, "该节点无有效审批人，自动通过"); err != nil {
				return err
			}
			continue
		}

		tasks := make([]models.WorkflowTask, 0, len(approvers))
		for _, u := range approvers {
			tasks = append(tasks, models.WorkflowTask{
				TenantID:     instance.TenantID,
				InstanceID:   instance.ID,
				NodeID:       node.ID,
				NodeName:     node.Name,
				NodeSort:     node.Sort,
				ApproverID:   u.ID,
				ApproverName: displayUserName(u),
				Status:       models.WorkflowTaskPending,
			})
		}
		if err := tx.Create(&tasks).Error; err != nil {
			return err
		}
		return tx.Model(&models.WorkflowInstance{}).Where("id = ?", instance.ID).
			Updates(map[string]interface{}{
				"current_sort": node.Sort,
				"current_node": node.Name,
			}).Error
	}

	// 已无后续节点：流程审批完成
	now := time.Now()
	if err := tx.Model(&models.WorkflowInstance{}).Where("id = ?", instance.ID).
		Updates(map[string]interface{}{
			"status":       models.WorkflowInstanceApproved,
			"current_sort": -1,
			"current_node": "已通过",
			"end_at":       now,
		}).Error; err != nil {
		return err
	}
	if err := createWorkflowLog(tx, instance, 0, "", 0, "系统", models.WorkflowActionFinish,
		"流程审批完成"); err != nil {
		return err
	}
	instance.Status = models.WorkflowInstanceApproved
	instance.CurrentSort = -1
	return nil
}

// resolveApprovers 解析节点审批人（仅启用状态用户）。
func resolveApprovers(tx *gorm.DB, node *models.WorkflowNode, initiatorID uint64) ([]models.User, error) {
	query := tx.Model(&models.User{}).Where("base_user.status = ?", 1)
	switch node.ApproverType {
	case models.ApproverTypeUser:
		query = query.Where("base_user.id = ?", node.ApproverID)
	case models.ApproverTypeInitiator:
		query = query.Where("base_user.id = ?", initiatorID)
	case models.ApproverTypeRole:
		query = query.Joins("JOIN base_workflow_role_user ON base_workflow_role_user.user_id = base_user.id").
			Where("base_workflow_role_user.workflow_role_id = ?", node.ApproverID)
	default:
		return nil, nil
	}
	var users []models.User
	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func loadWorkflowTask(tx *gorm.DB, taskID uint64) (*models.WorkflowTask, *models.WorkflowInstance, error) {
	var task models.WorkflowTask
	if err := tx.Where("id = ?", taskID).First(&task).Error; err != nil {
		return nil, nil, errors.New("审批任务不存在")
	}
	var instance models.WorkflowInstance
	if err := tx.Where("id = ?", task.InstanceID).First(&instance).Error; err != nil {
		return nil, nil, errors.New("流程实例不存在")
	}
	return &task, &instance, nil
}

func ensureApprovable(task *models.WorkflowTask, instance *models.WorkflowInstance, actor WorkflowActor) error {
	if task.ApproverID != actor.UserID {
		return errors.New("该审批任务不属于你")
	}
	if !models.IsPlatformTenant(actor.TenantID) && instance.TenantID != actor.TenantID {
		return errors.New("审批任务不存在")
	}
	if task.Status != models.WorkflowTaskPending {
		return errors.New("该任务已处理")
	}
	if instance.Status != models.WorkflowInstanceRunning {
		return errors.New("流程已结束，无法审批")
	}
	return nil
}

// invalidatePendingTasks 将实例下未处理的待办置为已失效（excludeTaskID 为本次已处理的节点任务）。
func invalidatePendingTasks(tx *gorm.DB, instanceID, excludeTaskID uint64, comment string, now time.Time) error {
	query := tx.Model(&models.WorkflowTask{}).
		Where("instance_id = ? AND status = ?", instanceID, models.WorkflowTaskPending)
	if excludeTaskID > 0 {
		query = query.Where("id <> ?", excludeTaskID)
	}
	return query.Updates(map[string]interface{}{
		"status":     models.WorkflowTaskInvalid,
		"comment":    comment,
		"handled_at": now,
	}).Error
}

func createWorkflowLog(tx *gorm.DB, instance *models.WorkflowInstance, nodeID uint64, nodeName string,
	operatorID uint64, operatorName, action, comment string) error {
	return tx.Create(&models.WorkflowLog{
		TenantID:     instance.TenantID,
		InstanceID:   instance.ID,
		NodeID:       nodeID,
		NodeName:     nodeName,
		OperatorID:   operatorID,
		OperatorName: operatorName,
		Action:       action,
		Comment:      comment,
	}).Error
}

// actorDisplayName 取操作人显示名（优先真实姓名），用于发起人/操作人快照。
func actorDisplayName(tx *gorm.DB, actor WorkflowActor) string {
	var u models.User
	if err := tx.Select("id", "username", "real_name").First(&u, actor.UserID).Error; err != nil {
		return actor.Username
	}
	return displayUserName(u)
}

func displayUserName(u models.User) string {
	if u.RealName != "" {
		return u.RealName
	}
	if u.Username != "" {
		return u.Username
	}
	return "未命名用户"
}
