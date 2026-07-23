package services

import (
	"errors"
	"slices"
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

func (s *OrganizationService) GetOrganizationByName(name string) (*models.Organization, error) {
	var org models.Organization
	if err := utils.DB.Where("name = ?", name).First(&org).Error; err != nil {
		return nil, errors.New("机构不存在")
	}
	return &org, nil
}

func (s *OrganizationService) CreateOrganization(org *models.Organization) error {
	return utils.DB.Create(org).Error
}

func (s *OrganizationService) UpdateOrganization(id uint, org *models.Organization) error {
	return utils.DB.Model(&models.Organization{}).Where("id = ?", id).UpdateColumns(map[string]any{
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
	return utils.DB.Model(&models.Organization{}).Where("id = ?", id).UpdateColumns(map[string]any{
		"user_ids":   strings.Join(ids, ","),
		"user_count": len(userIds),
	}).Error
}

func (s *OrganizationService) GetOrganizationByUserId(userId uint) (*models.Organization, error) {
	orgs, err := s.GetOrganizationsByUserId(userId)
	if err != nil {
		return nil, err
	}
	if len(orgs) == 0 {
		return nil, errors.New("未找到所属机构")
	}
	return &orgs[0], nil
}

func (s *OrganizationService) GetOrganizationsByUserId(userId uint) ([]models.Organization, error) {
	var orgs []models.Organization
	uidStr := strconv.Itoa(int(userId))
	err := utils.DB.Where("user_ids LIKE ? OR user_ids LIKE ? OR user_ids LIKE ?", "%"+uidStr+"%", "%"+uidStr+",%", "%,"+uidStr+"%").Find(&orgs).Error
	if err != nil {
		return nil, err
	}
	uid := int(userId)
	result := make([]models.Organization, 0)
	for _, org := range orgs {
		userIds, err := s.GetOrganizationUsers(org.ID)
		if err != nil {
			continue
		}
		if slices.Contains(userIds, uid) {
			result = append(result, org)
		}
	}
	return result, nil
}

func (s *OrganizationService) AddUserToOrganization(orgId uint, userId uint) error {
	if orgId == 0 {
		return nil
	}
	userIds, err := s.GetOrganizationUsers(orgId)
	if err != nil {
		return err
	}
	uid := int(userId)
	if slices.Contains(userIds, uid) {
		return nil
	}
	userIds = append(userIds, uid)
	return s.AssignOrganizationUsers(orgId, userIds)
}

func (s *OrganizationService) RemoveUserFromOrganization(orgId uint, userId uint) error {
	if orgId == 0 {
		return nil
	}
	userIds, err := s.GetOrganizationUsers(orgId)
	if err != nil {
		return nil
	}
	uid := int(userId)
	newUserIds := make([]int, 0, len(userIds))
	for _, id := range userIds {
		if id != uid {
			newUserIds = append(newUserIds, id)
		}
	}
	return s.AssignOrganizationUsers(orgId, newUserIds)
}

func (s *OrganizationService) RemoveUserFromAllOrganizations(userId uint) error {
	var orgs []models.Organization
	uidStr := strconv.Itoa(int(userId))
	err := utils.DB.Where("user_ids LIKE ? OR user_ids LIKE ? OR user_ids LIKE ?", "%"+uidStr+"%", "%"+uidStr+",%", "%,"+uidStr+"%").Find(&orgs).Error
	if err != nil {
		return err
	}
	for _, org := range orgs {
		if err := s.RemoveUserFromOrganization(org.ID, userId); err != nil {
			return err
		}
	}
	return nil
}
