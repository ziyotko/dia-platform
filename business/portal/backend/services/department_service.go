package services

import (
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"server/models"
	"server/utils"

	"github.com/xuri/excelize/v2"
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
	if err := validateTreeParent("department", 0, dept.ParentID); err != nil {
		return err
	}
	if err := ensureOrganizationExists(dept.OrgID); err != nil {
		return err
	}
	return utils.DB.Create(dept).Error
}

// ensureOrganizationExists 校验部门所属机构存在：
// 写入不存在的 org_id 会让部门在机构树/部门列表中都取不到机构名（脏数据，只能进库修正）。
func ensureOrganizationExists(orgID uint) error {
	if orgID == 0 {
		return nil
	}
	var count int64
	if err := utils.DB.Model(&models.Organization{}).Where("id = ?", orgID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("所属机构不存在")
	}
	return nil
}

type ImportDepartmentResult struct {
	SuccessCount int      `json:"successCount"`
	FailCount    int      `json:"failCount"`
	FailDetails  []string `json:"failDetails"`
}

func (s *DepartmentService) ImportDepartments(file multipart.File, fileSize int64) (*ImportDepartmentResult, error) {
	f, err := excelize.OpenReader(file, excelize.Options{})
	if err != nil {
		return nil, fmt.Errorf("读取 Excel 失败: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, errors.New("Excel 工作表为空")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取工作表失败: %w", err)
	}

	if len(rows) < 2 {
		return nil, errors.New("Excel 数据行数不足")
	}

	orgService := &OrganizationService{}
	result := &ImportDepartmentResult{
		SuccessCount: 0,
		FailCount:    0,
		FailDetails:  make([]string, 0),
	}

	for i, row := range rows[1:] {
		lineNum := i + 2
		if len(row) < 4 {
			result.FailCount++
			result.FailDetails = append(result.FailDetails, fmt.Sprintf("第 %d 行: 字段数量不足", lineNum))
			continue
		}

		parentIDStr := strings.TrimSpace(row[0])
		name := strings.TrimSpace(row[1])
		code := strings.TrimSpace(row[2])
		orgName := strings.TrimSpace(row[3])

		if name == "" || code == "" || orgName == "" {
			result.FailCount++
			result.FailDetails = append(result.FailDetails, fmt.Sprintf("第 %d 行: 部门名称、部门编码、机构名称不能为空", lineNum))
			continue
		}

		var parentID uint
		if parentIDStr != "" {
			pid, err := strconv.ParseUint(parentIDStr, 10, 32)
			if err != nil {
				result.FailCount++
				result.FailDetails = append(result.FailDetails, fmt.Sprintf("第 %d 行: 上级部门ID格式错误", lineNum))
				continue
			}
			parentID = uint(pid)
		}

		org, err := orgService.GetOrganizationByName(orgName)
		if err != nil {
			result.FailCount++
			result.FailDetails = append(result.FailDetails, fmt.Sprintf("第 %d 行: 机构 '%s' 不存在", lineNum, orgName))
			continue
		}

		dept := &models.Department{
			ParentID:    parentID,
			OrgID:       org.ID,
			Name:        name,
			Code:        code,
			Leader:      "",
			LeaderCode:  "",
			Sort:        1,
			Status:      1,
			Description: "",
		}

		if err := s.CreateDepartment(dept); err != nil {
			result.FailCount++
			result.FailDetails = append(result.FailDetails, fmt.Sprintf("第 %d 行 (%s): %s", lineNum, code, err.Error()))
			continue
		}

		result.SuccessCount++
	}

	return result, nil
}

func (s *DepartmentService) UpdateDepartment(id uint, dept *models.Department) error {
	if err := validateTreeParent("department", id, dept.ParentID); err != nil {
		return err
	}
	if err := ensureOrganizationExists(dept.OrgID); err != nil {
		return err
	}
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
	return utils.DB.Delete(&models.Department{}, id).Error
}

// parseUserIDs 将 department.user_ids（逗号分隔字符串）解析为 int 列表
func parseUserIDs(raw string) []int {
	ids := make([]int, 0)
	for _, uid := range parseMemberIDList(raw) {
		ids = append(ids, int(uid))
	}
	return ids
}

func (s *DepartmentService) GetDepartmentUsers(id uint) ([]int, error) {
	dept, err := s.GetDepartmentByID(id)
	if err != nil {
		return nil, err
	}
	return parseUserIDs(dept.UserIds), nil
}

// mutateDepartmentMembers 部门的「读-改-写」成员变更：
// 先对部门行加锁（SELECT ... FOR UPDATE），避免并发分配时相互覆盖（丢更新）；
// 落库前统一做去重、用户存在性校验与存储长度校验。
// rejectInvalid 为 true 时（覆盖式分配），请求中若包含不存在的用户则直接报错，避免“静默少了几个成员”。
func (s *DepartmentService) mutateDepartmentMembers(deptID uint, rejectInvalid bool, mutate func(existing []uint) []uint) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		var dept models.Department
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&dept, deptID).Error; err != nil {
			return errors.New("部门不存在")
		}
		kept, joined, dropped, err := normalizeMemberIDs(mutate(parseMemberIDList(dept.UserIds)), departmentMemberIDsMaxChars)
		if err != nil {
			return err
		}
		if rejectInvalid && dropped > 0 {
			return fmt.Errorf("有 %d 个成员不存在（可能已被删除），请刷新后重试", dropped)
		}
		return tx.Model(&models.Department{}).Where("id = ?", deptID).UpdateColumns(map[string]any{
			"user_ids":   joined,
			"user_count": len(kept),
		}).Error
	})
}

func (s *DepartmentService) AssignDepartmentUsers(id uint, userIds []int) error {
	return s.mutateDepartmentMembers(id, true, func(_ []uint) []uint {
		out := make([]uint, 0, len(userIds))
		for _, uid := range userIds {
			if uid > 0 {
				out = append(out, uint(uid))
			}
		}
		return out
	})
}

// RemoveUserFromAllDepartments 从所有部门中移除指定用户，并同步 user_count（删除用户时调用）。
func (s *DepartmentService) RemoveUserFromAllDepartments(userId uint) error {
	var departments []models.Department
	uidStr := strconv.Itoa(int(userId))
	err := utils.DB.Where("user_ids LIKE ? OR user_ids LIKE ? OR user_ids LIKE ?",
		"%"+uidStr+"%", "%"+uidStr+",%", "%,"+uidStr+"%").Find(&departments).Error
	if err != nil {
		return err
	}
	// 删除用户时的清理动作：单个部门失败不阻断其余部门
	for _, dept := range departments {
		if err := s.mutateDepartmentMembers(dept.ID, false, func(existing []uint) []uint {
			next := make([]uint, 0, len(existing))
			for _, id := range existing {
				if id != userId {
					next = append(next, id)
				}
			}
			return next
		}); err != nil {
			utils.Logger.Warnf("将用户[%d]从部门[%d]移除失败: %v", userId, dept.ID, err)
		}
	}
	return nil
}
