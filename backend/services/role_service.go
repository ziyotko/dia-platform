package services

import (
	"errors"
	"strconv"
	"strings"

	"server/models"
	"server/utils"
)

type RoleService struct{}

type RoleListResult struct {
	Total int64         `json:"total"`
	List  []models.Role `json:"list"`
}

func (s *RoleService) GetRoleList(page, pageSize int, name string) (*RoleListResult, error) {
	var roles []models.Role
	var total int64

	query := utils.DB.Model(&models.Role{})
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

	return &RoleListResult{
		Total: total,
		List:  roles,
	}, nil
}

func (s *RoleService) GetAllRoles() ([]models.Role, error) {
	var roles []models.Role
	err := utils.DB.Order("id ASC").Find(&roles).Error
	return roles, err
}

func (s *RoleService) GetRoleByID(id uint) (*models.Role, error) {
	var role models.Role
	if err := utils.DB.First(&role, id).Error; err != nil {
		return nil, errors.New("角色不存在")
	}
	return &role, nil
}

func (s *RoleService) CreateRole(role *models.Role) error {
	return utils.DB.Create(role).Error
}

func (s *RoleService) UpdateRole(id uint, role *models.Role) error {
	return utils.DB.Model(&models.Role{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":        role.Name,
		"code":        role.Code,
		"description": role.Description,
		"status":      role.Status,
		"permissions": role.Permissions,
	}).Error
}

func (s *RoleService) DeleteRole(id uint) error {
	return utils.DB.Unscoped().Delete(&models.Role{}, id).Error
}

func (s *RoleService) UpdateRolePermissions(id uint, permissions []uint) error {
	perms := make([]string, len(permissions))
	for i, p := range permissions {
		perms[i] = strconv.FormatUint(uint64(p), 10)
	}
	return utils.DB.Model(&models.Role{}).Where("id = ?", id).Update("permissions", strings.Join(perms, ",")).Error
}

func (s *RoleService) GetRolePermissions(id uint) ([]uint, error) {
	role, err := s.GetRoleByID(id)
	if err != nil {
		return nil, err
	}
	if role.Permissions == "" {
		return []uint{}, nil
	}
	parts := strings.Split(role.Permissions, ",")
	perms := make([]uint, 0, len(parts))
	for _, p := range parts {
		if val, err := strconv.ParseUint(strings.TrimSpace(p), 10, 32); err == nil {
			perms = append(perms, uint(val))
		}
	}
	return perms, nil
}
