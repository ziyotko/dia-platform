package services

import (
	"errors"
	"fmt"
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

// UpdateRole 更新角色基本信息。updatePermissions 为 false 时不写入 permissions：
// 角色编辑表单（名称/编码/描述/状态）不包含权限字段，若恒写会把已有菜单权限清空。
func (s *RoleService) UpdateRole(id uint, role *models.Role, updatePermissions bool) error {
	if models.IsBuiltinRoleID(id) {
		return errors.New("系统内置角色不允许修改，可在「分配权限」中调整其权限")
	}
	updates := map[string]any{
		"name":        role.Name,
		"code":        role.Code,
		"description": role.Description,
		"status":      role.Status,
	}
	if updatePermissions {
		updates["permissions"] = role.Permissions
	}
	return utils.DB.Model(&models.Role{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteRole 删除角色：仍被用户引用时拒绝。
// user.role_ids 是逗号分隔字符串（无外键），删除后该用户会残留一个不存在的角色 ID，
// GetUserMenus 解析不到其权限且错误被忽略 → 用户侧边栏静默清空、界面上无从自查。
func (s *RoleService) DeleteRole(id uint) error {
	if models.IsBuiltinRoleID(id) {
		return errors.New("系统内置角色不允许删除")
	}
	target := strconv.FormatUint(uint64(id), 10)
	var count int64
	if err := utils.DB.Model(&models.User{}).
		Where("role_ids = ? OR role_ids LIKE ? OR role_ids LIKE ? OR role_ids LIKE ?",
			target, target+",%", "%,"+target, "%,"+target+",%").
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("该角色仍被 %d 个用户使用，请先调整这些用户的角色后再删除", count)
	}
	return utils.DB.Delete(&models.Role{}, id).Error
}

func (s *RoleService) UpdateRolePermissions(id uint, permissions []uint) error {
	perms := make([]string, len(permissions))
	for i, p := range permissions {
		perms[i] = strconv.FormatUint(uint64(p), 10)
	}
	return utils.DB.Model(&models.Role{}).Where("id = ?", id).Update("permissions", strings.Join(perms, ",")).Error
}

// GetAllMenuIDs 返回全部菜单 ID（升序）。供「分配权限」展示管理员角色的“全部权限”：
// 管理员（角色 1）的权限不来自 permissions 字段（GetUserMenus 对角色 1 直接返回全部启用菜单），
// 若按库中空值下发，界面会显示「未勾选任何权限」，与实际权限不符。
func (s *RoleService) GetAllMenuIDs() ([]uint, error) {
	var ids []uint
	err := utils.DB.Model(&models.Menu{}).Order("id ASC").Pluck("id", &ids).Error
	return ids, err
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
