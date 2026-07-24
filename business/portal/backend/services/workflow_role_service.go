package services

import (
	"errors"
	"server/models"
	"server/utils"

	"gorm.io/gorm"
)

type WorkflowRoleService struct{}

type WorkflowRoleListResult struct {
	Total int64                 `json:"total"`
	List  []models.WorkflowRole `json:"list"`
}

func (s *WorkflowRoleService) GetWorkflowRoleList(page, pageSize int, name string) (*WorkflowRoleListResult, error) {
	var roles []models.WorkflowRole
	var total int64

	query := utils.DB.Model(&models.WorkflowRole{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	err = query.Order("id ASC").Offset(offset).Limit(pageSize).Find(&roles).Error
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
			records := make([]models.WorkflowRoleUser, 0, len(userIDs))
			for _, uid := range userIDs {
				records = append(records, models.WorkflowRoleUser{
					WorkflowRoleID: id,
					UserID:         uid,
				})
			}
			if err := tx.CreateInBatches(records, 100).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
