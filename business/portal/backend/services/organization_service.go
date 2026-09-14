package services

import (
	"errors"
	"fmt"
	"slices"
	"strconv"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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
	if err := validateTreeParent("organization", id, org.ParentID); err != nil {
		return err
	}
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
	if err := utils.DB.Model(&models.Organization{}).Where("parent_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("存在子机构，无法删除")
	}
	// 机构下仍有部门时不允许删除：否则 department.org_id 变成脏数据（组织名空白、部门在树中不可达）
	var deptCount int64
	if err := utils.DB.Model(&models.Department{}).Where("org_id = ?", id).Count(&deptCount).Error; err != nil {
		return err
	}
	if deptCount > 0 {
		return fmt.Errorf("该机构下存在 %d 个部门，请先移除或调整部门归属", deptCount)
	}
	return utils.DB.Unscoped().Delete(&models.Organization{}, id).Error
}

func (s *OrganizationService) GetOrganizationUsers(id uint) ([]int, error) {
	org, err := s.GetOrganizationByID(id)
	if err != nil {
		return nil, err
	}
	userIds := make([]int, 0)
	for _, uid := range parseMemberIDList(org.UserIds) {
		userIds = append(userIds, int(uid))
	}
	return userIds, nil
}

// mutateOrganizationMembers 机构的「读-改-写」成员变更：
// 先对机构行加锁（SELECT ... FOR UPDATE），避免并发分配时相互覆盖（丢更新）；
// 落库前统一做去重、用户存在性校验与存储长度校验。
// rejectInvalid 为 true 时（覆盖式分配），请求中若包含不存在的用户则直接报错，避免“静默少了几个成员”。
func (s *OrganizationService) mutateOrganizationMembers(orgID uint, rejectInvalid bool, mutate func(existing []uint) []uint) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		var org models.Organization
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&org, orgID).Error; err != nil {
			return errors.New("机构不存在")
		}
		kept, joined, dropped, err := normalizeMemberIDs(mutate(parseMemberIDList(org.UserIds)), organizationMemberIDsMaxChars)
		if err != nil {
			return err
		}
		if rejectInvalid && dropped > 0 {
			return fmt.Errorf("有 %d 个成员不存在（可能已被删除），请刷新后重试", dropped)
		}
		return tx.Model(&models.Organization{}).Where("id = ?", orgID).UpdateColumns(map[string]any{
			"user_ids":   joined,
			"user_count": len(kept),
		}).Error
	})
}

func (s *OrganizationService) AssignOrganizationUsers(id uint, userIds []int) error {
	return s.mutateOrganizationMembers(id, true, func(_ []uint) []uint {
		out := make([]uint, 0, len(userIds))
		for _, uid := range userIds {
			if uid > 0 {
				out = append(out, uint(uid))
			}
		}
		return out
	})
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
	if orgId == 0 || userId == 0 {
		return nil
	}
	return s.mutateOrganizationMembers(orgId, false, func(existing []uint) []uint {
		if slices.Contains(existing, userId) {
			return existing
		}
		return append(existing, userId)
	})
}

func (s *OrganizationService) RemoveUserFromOrganization(orgId uint, userId uint) error {
	if orgId == 0 {
		return nil
	}
	return s.mutateOrganizationMembers(orgId, false, func(existing []uint) []uint {
		next := make([]uint, 0, len(existing))
		for _, id := range existing {
			if id != userId {
				next = append(next, id)
			}
		}
		return next
	})
}

func (s *OrganizationService) RemoveUserFromAllOrganizations(userId uint) error {
	var orgs []models.Organization
	uidStr := strconv.Itoa(int(userId))
	err := utils.DB.Where("user_ids LIKE ? OR user_ids LIKE ? OR user_ids LIKE ?", "%"+uidStr+"%", "%"+uidStr+",%", "%,"+uidStr+"%").Find(&orgs).Error
	if err != nil {
		return err
	}
	// 删除用户时的清理动作：单个机构失败不阻断其余机构
	for _, org := range orgs {
		if err := s.RemoveUserFromOrganization(org.ID, userId); err != nil {
			utils.Logger.Warnf("将用户[%d]从机构[%d]移除失败: %v", userId, org.ID, err)
		}
	}
	return nil
}
