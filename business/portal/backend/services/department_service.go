package services

import (
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"

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
	return utils.DB.Create(dept).Error
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
