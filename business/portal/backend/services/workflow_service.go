package services

import (
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
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	err = query.Preload("Nodes").Order("id DESC").Limit(pageSize).Offset(offset).Find(&workflows).Error
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
	return utils.DB.Unscoped().Delete(&models.Workflow{}, id).Error
}

func (s *WorkflowService) SaveWorkflowNodes(workflowID uint, nodes []models.WorkflowNode) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("workflow_id = ?", workflowID).Unscoped().Delete(&models.WorkflowNode{}).Error; err != nil {
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
