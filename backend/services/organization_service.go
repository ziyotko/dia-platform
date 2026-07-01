package services

import (
	"errors"
	"strconv"
	"strings"

	"server/models"
	"server/utils"
)

type OrganizationService struct{}

type OrganizationListResult struct {
	Total int64                 `json:"total"`
	List  []models.Organization `json:"list"`
}

func (s *OrganizationService) GetOrganizationList(name string, orgType *int, status *int) (*OrganizationListResult, error) {
	var organizations []models.Organization
	var total int64

	query := utils.DB.Model(&models.Organization{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if orgType != nil {
		query = query.Where("org_type = ?", *orgType)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	err = query.Order("sort ASC, id ASC").Find(&organizations).Error
	if err != nil {
		return nil, err
	}

	return &OrganizationListResult{
		Total: total,
		List:  organizations,
	}, nil
}

func (s *OrganizationService) GetOrganizationTree() ([]models.Organization, error) {
	var organizations []models.Organization
	err := utils.DB.Order("sort ASC, id ASC").Find(&organizations).Error
	if err != nil {
		return nil, err
	}
	return organizations, nil
}

func (s *OrganizationService) GetOrganizationByID(id uint) (*models.Organization, error) {
	var org models.Organization
	if err := utils.DB.First(&org, id).Error; err != nil {
		return nil, errors.New("机构不存在")
	}
	return &org, nil
}

func (s *OrganizationService) CreateOrganization(org *models.Organization) error {
	return utils.DB.Create(org).Error
}

func (s *OrganizationService) UpdateOrganization(id uint, org *models.Organization) error {
	return utils.DB.Model(&models.Organization{}).Where("id = ?", id).UpdateColumns(map[string]interface{}{
		"parent_id":    org.ParentID,
		"name":         org.Name,
		"code":         org.Code,
		"org_type":     org.OrgType,
		"org_level":    org.OrgLevel,
		"category":     org.Category,
		"region":       org.Region,
		"province":     org.Province,
		"city":         org.City,
		"address":      org.Address,
		"manager":      org.Manager,
		"manager_code": org.ManagerCode,
		"sort":         org.Sort,
		"status":       org.Status,
		"description":  org.Description,
	}).Error
}

func (s *OrganizationService) DeleteOrganization(id uint) error {
	var count int64
	utils.DB.Model(&models.Organization{}).Where("parent_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("存在子机构，无法删除")
	}
	return utils.DB.Unscoped().Delete(&models.Organization{}, id).Error
}

func (s *OrganizationService) GetOrganizationUsers(id uint) ([]int, error) {
	org, err := s.GetOrganizationByID(id)
	if err != nil {
		return nil, err
	}
	if org.UserIds == "" {
		return []int{}, nil
	}
	parts := strings.Split(org.UserIds, ",")
	userIds := make([]int, 0, len(parts))
	for _, p := range parts {
		if val, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
			userIds = append(userIds, val)
		}
	}
	return userIds, nil
}

func (s *OrganizationService) AssignOrganizationUsers(id uint, userIds []int) error {
	ids := make([]string, len(userIds))
	for i, uid := range userIds {
		ids[i] = strconv.Itoa(uid)
	}
	return utils.DB.Model(&models.Organization{}).Where("id = ?", id).UpdateColumns(map[string]interface{}{
		"user_ids":   strings.Join(ids, ","),
		"user_count": len(userIds),
	}).Error
}
