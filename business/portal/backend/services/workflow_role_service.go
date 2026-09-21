package services

import (
	"errors"
	"fmt"
	"server/models"
	"server/utils"

	"gorm.io/gorm"
)

type WorkflowRoleService struct{}

type WorkflowRoleListResult struct {
	Total int64                 `json:"total"`
	List  []models.WorkflowRole `json:"list"`
}

func (s *WorkflowRoleService) GetWorkflowRoleList(page, pageSize int, name string, status *int) (*WorkflowRoleListResult, error) {
	var roles []models.WorkflowRole
	var total int64

	query := utils.DB.Model(&models.WorkflowRole{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	// pageSize <= 0 表示不限制（供下拉选项使用，避免被分页截断）
	listQuery := query.Order("id ASC")
	if pageSize > 0 {
		listQuery = listQuery.Offset((page - 1) * pageSize).Limit(pageSize)
	}
	err = listQuery.Find(&roles).Error
	if err != nil {
		return nil, err
	}

	return &WorkflowRoleListResult{
		Total: total,
		List:  roles,
	}, nil
}

func (s *WorkflowRoleService) GetWorkflowRoleByID(id uint) (*models.WorkflowRole, error) {
	var role models.WorkflowRole
	if err := utils.DB.First(&role, id).Error; err != nil {
		return nil, errors.New("流程角色不存在")
	}
	return &role, nil
}

func (s *WorkflowRoleService) CreateWorkflowRole(role *models.WorkflowRole) error {
	return utils.DB.Create(role).Error
}

func (s *WorkflowRoleService) UpdateWorkflowRole(id uint, role *models.WorkflowRole) error {
	return utils.DB.Model(&models.WorkflowRole{}).Where("id = ?", id).Updates(map[string]any{
		"name":        role.Name,
		"code":        role.Code,
		"description": role.Description,
		"status":      role.Status,
	}).Error
}

func (s *WorkflowRoleService) DeleteWorkflowRole(id uint) error {
	// 被审批节点引用时不允许删除：删除后这些节点将无人可审（approver_id 永远匹配不到任何用户）
	var nodeCount int64
	if err := utils.DB.Model(&models.WorkflowNode{}).
		Where("approver_type = ? AND approver_id = ?", "role", id).
		Count(&nodeCount).Error; err != nil {
		return err
	}
	if nodeCount > 0 {
		return fmt.Errorf("该流程角色已被 %d 个审批节点使用，请先调整流程节点", nodeCount)
	}
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("workflow_role_id = ?", id).Delete(&models.WorkflowRoleUser{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Delete(&models.WorkflowRole{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *WorkflowRoleService) GetWorkflowRoleUsers(id uint) ([]models.User, error) {
	var users []models.User
	err := utils.DB.Model(&models.User{}).
		Joins("JOIN workflow_role_user ON workflow_role_user.user_id = user.id").
		Where("workflow_role_user.workflow_role_id = ?", id).
		Find(&users).Error
	return users, err
}

func (s *WorkflowRoleService) GetUserWorkflowRoleIds(userID uint) ([]uint, error) {
	var roleIDs []uint
	err := utils.DB.Model(&models.WorkflowRoleUser{}).
		Select("workflow_role_id").
		Where("user_id = ?", userID).
		Scan(&roleIDs).Error
	return roleIDs, err
}

func (s *WorkflowRoleService) UpdateWorkflowRoleUsers(id uint, userIDs []uint) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("workflow_role_id = ?", id).Delete(&models.WorkflowRoleUser{}).Error; err != nil {
			return err
		}
		if len(userIDs) > 0 {
			// 去重并剔除不存在的用户：成员来源为前端多选，重复提交会产生重复行，
			// 已删除用户残留则会产生悬挂成员（审批人解析时看似有成员实则无人）。
			var existing []uint
			if err := tx.Model(&models.User{}).Where("id IN ?", userIDs).Pluck("id", &existing).Error; err != nil {
				return err
			}
			valid := make(map[uint]bool, len(existing))
			for _, uid := range existing {
				valid[uid] = true
			}
			records := make([]models.WorkflowRoleUser, 0, len(userIDs))
			seen := make(map[uint]bool, len(userIDs))
			for _, uid := range userIDs {
				if uid == 0 || seen[uid] || !valid[uid] {
					continue
				}
				seen[uid] = true
				records = append(records, models.WorkflowRoleUser{
					WorkflowRoleID: id,
					UserID:         uid,
				})
			}
			if len(records) > 0 {
				if err := tx.CreateInBatches(records, 100).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
