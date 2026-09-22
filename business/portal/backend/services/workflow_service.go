package services

import (
	"errors"
	"fmt"

	"server/models"
	"server/utils"

	"gorm.io/gorm"
)

type WorkflowService struct{}

func (s *WorkflowService) GetWorkflows(name string, page int, pageSize int) ([]models.Workflow, int64, error) {
	var workflows []models.Workflow
	var total int64
	query := utils.DB.Model(&models.Workflow{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	// pageSize <= 0 表示不限制（供下拉选项使用，避免被分页截断）
	listQuery := query.Preload("Nodes").Order("id DESC")
	if pageSize > 0 {
		listQuery = listQuery.Limit(pageSize).Offset((page - 1) * pageSize)
	}
	err = listQuery.Find(&workflows).Error
	return workflows, total, err
}

func (s *WorkflowService) GetWorkflowByID(id uint) (*models.Workflow, error) {
	var workflow models.Workflow
	err := utils.DB.Preload("Nodes").First(&workflow, id).Error
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (s *WorkflowService) CreateWorkflow(workflow *models.Workflow) error {
	return utils.DB.Create(workflow).Error
}

func (s *WorkflowService) UpdateWorkflow(id uint, workflow *models.Workflow) error {
	var old models.Workflow
	if err := utils.DB.First(&old, id).Error; err != nil {
		return err
	}
	return utils.DB.Model(&old).Updates(map[string]any{
		"name":        workflow.Name,
		"status":      workflow.Status,
		"description": workflow.Description,
	}).Error
}

func (s *WorkflowService) DeleteWorkflow(id uint) error {
	// 仍被栏目绑定：删除后 column.workflow_id 会变成脏数据，该栏目提交审核时会被当成「无流程」直接通过
	var columnCount int64
	if err := utils.DB.Model(&models.Column{}).Where("workflow_id = ?", id).Count(&columnCount).Error; err != nil {
		return err
	}
	if columnCount > 0 {
		return fmt.Errorf("该流程已被 %d 个栏目绑定，请先解除绑定后再删除", columnCount)
	}
	// 存在进行中的审核：删除后这些审核无法继续
	var pendingCount int64
	if err := utils.DB.Model(&models.ArticleColumnAudit{}).
		Where("workflow_id = ? AND status = ?", id, 0).
		Count(&pendingCount).Error; err != nil {
		return err
	}
	if pendingCount > 0 {
		return fmt.Errorf("该流程存在 %d 条进行中的审核，无法删除", pendingCount)
	}
	return utils.DB.Delete(&models.Workflow{}, id).Error
}

// ValidateNodesResolvable 校验节点审批人配置是否可解析。
// 审批人为空的节点会让审核永久卡住（且历史上等价于「人人可审」），因此在保存时与发起审核时都要校验。
func (s *WorkflowService) ValidateNodesResolvable(nodes []models.WorkflowNode) error {
	for _, node := range nodes {
		approverType := node.ApproverType
		if approverType == "" {
			approverType = "user"
		}
		switch approverType {
		case "user":
			if node.ApproverID == 0 {
				return fmt.Errorf("节点「%s」未指定审批人", node.Name)
			}
			var count int64
			if err := utils.DB.Model(&models.User{}).Where("id = ?", node.ApproverID).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				return fmt.Errorf("节点「%s」指定的审批人不存在", node.Name)
			}
		case "role":
			if node.ApproverID == 0 {
				return fmt.Errorf("节点「%s」未指定审批角色", node.Name)
			}
			var count int64
			if err := utils.DB.Model(&models.WorkflowRole{}).Where("id = ?", node.ApproverID).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				return fmt.Errorf("节点「%s」指定的审批角色不存在", node.Name)
			}
		case "dept_head":
			// 部门负责人按作者所在部门动态解析，无需指定审批人
		default:
			return fmt.Errorf("节点「%s」的审批人类型无效: %s", node.Name, node.ApproverType)
		}
	}
	return nil
}

// ValidateWorkflowResolvable 校验指定流程是否可用于送审（流程处于启用状态 + 至少一个节点 + 审批人可解析）。
// 仅由「提交审核」调用；已禁用的流程不应再被送审使用（前端下拉只列启用流程，
// 但「先绑定后禁用」或直调 API 仍会遇到，原先会静默按禁用流程送审）。
func (s *WorkflowService) ValidateWorkflowResolvable(workflowID uint) error {
	var workflow models.Workflow
	if err := utils.DB.First(&workflow, workflowID).Error; err != nil {
		return fmt.Errorf("审核流程不存在或已被删除")
	}
	if workflow.Status != 1 {
		return fmt.Errorf("审核流程「%s」已禁用，请先启用该流程或为栏目改绑其它流程", workflow.Name)
	}
	nodes, err := s.GetWorkflowNodes(workflowID)
	if err != nil {
		return err
	}
	if len(nodes) == 0 {
		return errors.New("该审核流程未配置审批节点")
	}
	return s.ValidateNodesResolvable(nodes)
}

func (s *WorkflowService) SaveWorkflowNodes(workflowID uint, nodes []models.WorkflowNode) error {
	// 先校验再落库：校验失败时不破坏已存在的节点配置
	if err := s.ValidateNodesResolvable(nodes); err != nil {
		return err
	}
	// 节点排序值必须唯一：审核推进是按 (sort_order, id) 顺序取下一节点，
	// 重复值会让审批顺序变得不确定（历史上更会因「找不到下一个节点」而直接判定审核通过→自动发布）。
	if err := validateNodeSortOrders(nodes); err != nil {
		return err
	}
	// 节点是「全删重建」（ID 会变化），若存在进行中的审核，其 current_node_id 将指向已删节点而永久卡住
	var pendingCount int64
	if err := utils.DB.Model(&models.ArticleColumnAudit{}).
		Where("workflow_id = ? AND status = ?", workflowID, 0).
		Count(&pendingCount).Error; err != nil {
		return err
	}
	if pendingCount > 0 {
		return fmt.Errorf("该流程存在 %d 条进行中的审核，请等待审核完成或先撤回后再修改节点", pendingCount)
	}
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("workflow_id = ?", workflowID).Delete(&models.WorkflowNode{}).Error; err != nil {
			return err
		}
		if len(nodes) > 0 {
			for i := range nodes {
				nodes[i].WorkflowID = workflowID
				nodes[i].ID = 0
			}
			if err := tx.Create(&nodes).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// validateNodeSortOrders 校验节点排序值唯一（同一流程内不允许重复）。
func validateNodeSortOrders(nodes []models.WorkflowNode) error {
	seen := make(map[int]bool, len(nodes))
	for _, node := range nodes {
		if seen[node.SortOrder] {
			return fmt.Errorf("存在重复的节点排序值（%d），请为每个节点设置唯一的排序", node.SortOrder)
		}
		seen[node.SortOrder] = true
	}
	return nil
}

func (s *WorkflowService) GetWorkflowNodes(workflowID uint) ([]models.WorkflowNode, error) {
	var nodes []models.WorkflowNode
	err := utils.DB.Where("workflow_id = ?", workflowID).Order("sort_order ASC").Find(&nodes).Error
	return nodes, err
}

func (s *WorkflowService) GetWorkflowNodeByID(id uint) (*models.WorkflowNode, error) {
	var node models.WorkflowNode
	err := utils.DB.First(&node, id).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}
