package services

import (
	"errors"
	"strconv"
	"strings"

	"server/models"
	"server/utils"
)

type DepartmentService struct{}

type DepartmentListResult struct {
	Total int64               `json:"total"`
	List  []models.Department `json:"list"`
}

func (s *DepartmentService) GetDepartmentList(name string, status *int) (*DepartmentListResult, error) {
	var departments []models.Department
	var total int64

	query := utils.DB.Model(&models.Department{})
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

	err = query.Order("sort ASC, id ASC").Find(&departments).Error
	if err != nil {
		return nil, err
	}

	return &DepartmentListResult{
		Total: total,
		List:  departments,
	}, nil
}

func (s *DepartmentService) GetDepartmentTree() ([]models.Department, error) {
	var departments []models.Department
	err := utils.DB.Order("sort ASC, id ASC").Find(&departments).Error
	if err != nil {
		return nil, err
	}
	return departments, nil
}

func (s *DepartmentService) GetDepartmentsByOrgID(orgID uint) ([]models.Department, error) {
	var departments []models.Department
	err := utils.DB.Where("org_id = ?", orgID).Order("sort ASC, id ASC").Find(&departments).Error
	if err != nil {
		return nil, err
	}
	return departments, nil
}

func (s *DepartmentService) GetDepartmentByID(id uint) (*models.Department, error) {
	var dept models.Department
	if err := utils.DB.First(&dept, id).Error; err != nil {
		return nil, errors.New("部门不存在")
	}
	return &dept, nil
}

func (s *DepartmentService) CreateDepartment(dept *models.Department) error {
	return utils.DB.Create(dept).Error
}

func (s *DepartmentService) UpdateDepartment(id uint, dept *models.Department) error {
	return utils.DB.Model(&models.Department{}).Where("id = ?", id).UpdateColumns(map[string]any{
		"parent_id":   dept.ParentID,
		"org_id":      dept.OrgID,
		"name":        dept.Name,
		"code":        dept.Code,
		"leader":      dept.Leader,
		"leader_code": dept.LeaderCode,
		"sort":        dept.Sort,
		"status":      dept.Status,
		"description": dept.Description,
	}).Error
}

func (s *DepartmentService) DeleteDepartment(id uint) error {
	var count int64
	utils.DB.Model(&models.Department{}).Where("parent_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("存在子部门，无法删除")
	}
	return utils.DB.Unscoped().Delete(&models.Department{}, id).Error
}

func (s *DepartmentService) GetDepartmentUsers(id uint) ([]int, error) {
	dept, err := s.GetDepartmentByID(id)
	if err != nil {
		return nil, err
	}
	if dept.UserIds == "" {
		return []int{}, nil
	}
	parts := strings.Split(dept.UserIds, ",")
	userIds := make([]int, 0, len(parts))
	for _, p := range parts {
		if val, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
			userIds = append(userIds, val)
		}
	}
	return userIds, nil
}

func (s *DepartmentService) AssignDepartmentUsers(id uint, userIds []int) error {
	ids := make([]string, len(userIds))
	for i, uid := range userIds {
		ids[i] = strconv.Itoa(uid)
	}
	return utils.DB.Model(&models.Department{}).Where("id = ?", id).UpdateColumns(map[string]any{
		"user_ids":   strings.Join(ids, ","),
		"user_count": len(userIds),
	}).Error
}
