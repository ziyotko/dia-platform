package service

import (
	"errors"
	"fmt"
	"time"

	"base/internal/models"
	"base/pkg/db"

	"github.com/sirupsen/logrus"
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
		if err := createWorkflowLog(tx, instance, task.NodeID, task.NodeName, actor.UserID, name, models.WorkflowActionApprove, comment); err != nil {
			return err
		}

		if task.ApproveMode == models.ApproveModeAnd {
			// 会签：同节点还有其他待处理任务时停在当前节点，全部通过才推进
			var pending int64
			if err := tx.Model(&models.WorkflowTask{}).
				Where("instance_id = ? AND node_id = ? AND status = ?", instance.ID, task.NodeID, models.WorkflowTaskPending).
				Count(&pending).Error; err != nil {
				return err
			}
			if pending > 0 {
				return nil
			}
			return advanceWorkflow(tx, instance, task.NodeSort)
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

// ApproverOptions 转办/加签可选的用户：本租户启用中的用户（平台超管可见全部）。
// 与流程定义里的审批人候选不同，这里只返回用户，不需要流程定义的管理权限。
func (s WorkflowEngineService) ApproverOptions(actor WorkflowActor) ([]WorkflowUserOption, error) {
	query := db.DB.Model(&models.User{}).Where("status = ?", 1).Order("username ASC")
	if !models.IsPlatformTenant(actor.TenantID) {
		query = query.Where("tenant_id = ?", actor.TenantID)
	}
	var users []models.User
	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}
	options := make([]WorkflowUserOption, 0, len(users))
	for _, u := range users {
		options = append(options, WorkflowUserOption{
			ID:       u.ID,
			Username: u.Username,
			RealName: u.RealName,
			Status:   u.Status,
		})
	}
	return options, nil
}

// TransferRequest 转办请求
type TransferRequest struct {
	UserID  uint64 `json:"userId"`
	Comment string `json:"comment"`
}

// AddApproverRequest 加签请求
type AddApproverRequest struct {
	UserID  uint64 `json:"userId"`
	Comment string `json:"comment"`
}

// Transfer 转办：把待办交给同租户的另一个用户，原审批人不再需要处理。
func (s WorkflowEngineService) Transfer(taskID uint64, req TransferRequest, actor WorkflowActor) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		task, instance, err := loadWorkflowTask(tx, taskID)
		if err != nil {
			return err
		}
		if err := ensureApprovable(task, instance, actor); err != nil {
			return err
		}
		if req.UserID == 0 {
			return errors.New("请选择转办人")
		}
		if req.UserID == actor.UserID {
			return errors.New("不能转办给自己")
		}
		target, err := findWorkflowUser(tx, instance.TenantID, req.UserID)
		if err != nil {
			return err
		}
		if err := ensureNoDuplicateTask(tx, instance.ID, task.NodeID, target.ID, task.ID); err != nil {
			return err
		}
		name := actorDisplayName(tx, actor)
		if err := tx.Model(&models.WorkflowTask{}).Where("id = ?", task.ID).
			Updates(map[string]interface{}{
				"approver_id":   target.ID,
				"approver_name": displayUserName(*target),
			}).Error; err != nil {
			return err
		}
		return createWorkflowLog(tx, instance, task.NodeID, task.NodeName, actor.UserID, name,
			models.WorkflowActionTransfer, summarizeAction(fmt.Sprintf("%s 转办给 %s", name, displayUserName(*target)), req.Comment))
	})
}

// AddApprover 加签：在当前节点追加一个审批人（会签节点最常用；或签节点下新加的人也能处理该节点）。
func (s WorkflowEngineService) AddApprover(taskID uint64, req AddApproverRequest, actor WorkflowActor) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		task, instance, err := loadWorkflowTask(tx, taskID)
		if err != nil {
			return err
		}
		if err := ensureApprovable(task, instance, actor); err != nil {
			return err
		}
		if req.UserID == 0 {
			return errors.New("请选择加签人")
		}
		target, err := findWorkflowUser(tx, instance.TenantID, req.UserID)
		if err != nil {
			return err
		}
		if err := ensureNoDuplicateTask(tx, instance.ID, task.NodeID, target.ID, 0); err != nil {
			return err
		}
		extra := models.WorkflowTask{
			TenantID:       instance.TenantID,
			InstanceID:     instance.ID,
			NodeID:         task.NodeID,
			NodeName:       task.NodeName,
			NodeSort:       task.NodeSort,
			ApproveMode:    normalizeApproveMode(task.ApproveMode),
			TimeoutMinutes: task.TimeoutMinutes,
			ApproverID:     target.ID,
			ApproverName:   displayUserName(*target),
			Status:         models.WorkflowTaskPending,
		}
		if err := tx.Create(&extra).Error; err != nil {
			return err
		}
		name := actorDisplayName(tx, actor)
		return createWorkflowLog(tx, instance, task.NodeID, task.NodeName, actor.UserID, name,
			models.WorkflowActionAddApprover, summarizeAction(fmt.Sprintf("%s 加签 %s", name, displayUserName(*target)), req.Comment))
	})
}

// findWorkflowUser 校验转办/加签的目标用户：必须启用，且与实例同租户（平台级实例不限租户）。
func findWorkflowUser(tx *gorm.DB, instanceTenantID, userID uint64) (*models.User, error) {
	var user models.User
	if err := tx.Where("id = ? AND status = ?", userID, 1).First(&user).Error; err != nil {
		return nil, errors.New("所选用户不存在或已停用")
	}
	if instanceTenantID > 0 && user.TenantID != instanceTenantID {
		return nil, errors.New("所选用户不属于该流程所在租户")
	}
	return &user, nil
}

// ensureNoDuplicateTask 同一实例同一节点不允许同一审批人重复出现（excludeTaskID 为当前任务）。
func ensureNoDuplicateTask(tx *gorm.DB, instanceID, nodeID, userID, excludeTaskID uint64) error {
	query := tx.Model(&models.WorkflowTask{}).
		Where("instance_id = ? AND node_id = ? AND approver_id = ? AND status = ?",
			instanceID, nodeID, userID, models.WorkflowTaskPending)
	if excludeTaskID > 0 {
		query = query.Where("id <> ?", excludeTaskID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该用户已是本节点的待处理审批人")
	}
	return nil
}

// summarizeAction 把操作摘要与用户填写的意见拼成一条日志备注（Comment 字段长度有限）。
func summarizeAction(summary, comment string) string {
	if comment == "" {
		return summary
	}
	runes := []rune(comment)
	if len(runes) > 200 {
		comment = string(runes[:200]) + "..."
	}
	return summary + "：" + comment
}

// RemindOverdueTasks 扫描超时未处理的待办，向审批人发站内信催办并写流转日志，返回本轮催办数量。
// 每个「超时周期」最多提醒一次（依赖 reminded_at，不会反复骚扰）。
func (s WorkflowEngineService) RemindOverdueTasks() (int, error) {
	var tasks []models.WorkflowTask
	if err := db.DB.Preload("Instance").
		Joins("JOIN base_workflow_instance ON base_workflow_instance.id = base_workflow_task.instance_id").
		Where("base_workflow_task.status = ?", models.WorkflowTaskPending).
		Where("base_workflow_instance.status = ?", models.WorkflowInstanceRunning).
		Where("base_workflow_task.timeout_minutes > 0").
		Where("TIMESTAMPADD(MINUTE, base_workflow_task.timeout_minutes, base_workflow_task.created_at) <= NOW()").
		Where("base_workflow_task.reminded_at IS NULL OR TIMESTAMPADD(MINUTE, base_workflow_task.timeout_minutes, base_workflow_task.reminded_at) <= NOW()").
		Order("base_workflow_task.id ASC").
		Limit(200).
		Find(&tasks).Error; err != nil {
		return 0, err
	}
	if len(tasks) == 0 {
		return 0, nil
	}

	msgSvc := MessageService{}
	reminded := 0
	for i := range tasks {
		task := tasks[i]
		instanceTitle := fmt.Sprintf("流程实例 #%d", task.InstanceID)
		if task.Instance != nil && task.Instance.Title != "" {
			instanceTitle = task.Instance.Title
		}
		content := fmt.Sprintf("你的待办已超过 %d 分钟未处理：%s / %s，请尽快处理。",
			task.TimeoutMinutes, instanceTitle, task.NodeName)

		// 站内信发送失败（例如审批人已被停用）不影响其它待办的催办
		if err := msgSvc.SendToUsers(0, "系统", task.TenantID, []uint64{task.ApproverID},
			"审批超时提醒", content, "system", "high"); err != nil {
			logrus.WithError(err).Warnf("工作流超时提醒发送失败: taskID=%d", task.ID)
		}

		now := time.Now()
		if err := db.DB.Model(&models.WorkflowTask{}).Where("id = ?", task.ID).
			Updates(map[string]interface{}{
				"reminded_at":  now,
				"remind_count": gorm.Expr("remind_count + 1"),
			}).Error; err != nil {
			return reminded, err
		}
		if err := db.DB.Create(&models.WorkflowLog{
			TenantID:     task.TenantID,
			InstanceID:   task.InstanceID,
			NodeID:       task.NodeID,
			NodeName:     task.NodeName,
			OperatorID:   0,
			OperatorName: "系统",
			Action:       models.WorkflowActionRemind,
			Comment:      content,
		}).Error; err != nil {
			return reminded, err
		}
		reminded++
	}
	return reminded, nil
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
				TenantID:       instance.TenantID,
				InstanceID:     instance.ID,
				NodeID:         node.ID,
				NodeName:       node.Name,
				NodeSort:       node.Sort,
				ApproveMode:    normalizeApproveMode(node.ApproveMode),
				TimeoutMinutes: node.TimeoutMinutes,
				ApproverID:     u.ID,
				ApproverName:   displayUserName(u),
				Status:         models.WorkflowTaskPending,
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

// normalizeApproveMode 归一化审批方式（历史数据/空值按或签处理）。
func normalizeApproveMode(mode string) string {
	if mode == models.ApproveModeAnd {
		return models.ApproveModeAnd
	}
	return models.ApproveModeOr
}

// loadWorkflowTask 读取任务及其所属实例。
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
