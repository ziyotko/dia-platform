package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/services"
	"server/utils"
)

type DepartmentController struct {
	deptService *services.DepartmentService
	orgService  *services.OrganizationService
}

func NewDepartmentController() *DepartmentController {
	return &DepartmentController{
		deptService: &services.DepartmentService{},
		orgService:  &services.OrganizationService{},
	}
}

func buildOrgNameMap(orgs []models.Organization) map[uint]string {
	m := make(map[uint]string, len(orgs))
	for i := range orgs {
		m[orgs[i].ID] = orgs[i].Name
	}
	return m
}

func buildDeptTree(list []models.Department, orgNameMap map[uint]string) []gin.H {
	return buildFlatTree(list,
		func(item models.Department) uint { return item.ID },
		func(item models.Department) uint { return item.ParentID },
		func(item models.Department) gin.H {
			orgName := ""
			if item.OrgID > 0 {
				orgName = orgNameMap[item.OrgID]
			}
			return gin.H{
				"id":          item.ID,
				"parentId":    item.ParentID,
				"orgId":       item.OrgID,
				"orgName":     orgName,
				"name":        item.Name,
				"code":        item.Code,
				"leader":      item.Leader,
				"leaderCode":  item.LeaderCode,
				"sort":        item.Sort,
				"status":      item.Status,
				"description": item.Description,
				"userCount":   item.UserCount,
				"createdAt":   item.CreatedAt.Format("2006-01-02 15:04:05"),
				"children":    []gin.H{},
				"hasChildren": false,
			}
		})
}

func (c *DepartmentController) GetDepartments(ctx *gin.Context) {
	name := ctx.Query("name")
	statusStr := ctx.Query("status")

	var status *int
	if statusStr != "" {
		s, _ := strconv.Atoi(statusStr)
		status = &s
	}

	result, err := c.deptService.GetDepartmentList(name, status)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取部门列表失败"))
		return
	}

	orgs, _ := c.orgService.GetOrganizationTree()
	tree := buildDeptTree(result.List, buildOrgNameMap(orgs))
	ctx.JSON(http.StatusOK, utils.Success("获取部门列表成功", utils.AllData(tree, result.Total)))
}

func (c *DepartmentController) GetDepartmentTree(ctx *gin.Context) {
	list, err := c.deptService.GetDepartmentTree()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取部门树失败"))
		return
	}
	orgs, _ := c.orgService.GetOrganizationTree()
	tree := buildDeptTree(list, buildOrgNameMap(orgs))
	ctx.JSON(http.StatusOK, utils.Success("获取部门树成功", tree))
}

func (c *DepartmentController) CreateDepartment(ctx *gin.Context) {
	var req models.Department
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}
	if err := c.deptService.CreateDepartment(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("创建部门失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("创建部门成功", nil))
}

func (c *DepartmentController) ImportDepartments(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("上传文件失败", err)))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("打开文件失败", err)))
		return
	}
	defer file.Close()

	result, err := c.deptService.ImportDepartments(file, fileHeader.Size)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("导入失败", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("导入完成", result))
}

func (c *DepartmentController) UpdateDepartment(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "部门ID无效"))
		return
	}

	var req models.Department
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}

	if err := c.deptService.UpdateDepartment(uint(id), &req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("更新部门失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新部门成功", nil))
}

func (c *DepartmentController) DeleteDepartment(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "部门ID无效"))
		return
	}

	if err := c.deptService.DeleteDepartment(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("删除部门失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("删除部门成功", nil))
}

func (c *DepartmentController) GetDepartmentUsers(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "部门ID无效"))
		return
	}

	userIds, err := c.deptService.GetDepartmentUsers(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SafeErrText(err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取部门用户成功", userIds))
}

func (c *DepartmentController) AssignDepartmentUsers(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "部门ID无效"))
		return
	}

	var req struct {
		UserIds []int `json:"userIds"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误"))
		return
	}

	if err := c.deptService.AssignDepartmentUsers(uint(id), req.UserIds); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("分配用户失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("分配用户成功", nil))
}
