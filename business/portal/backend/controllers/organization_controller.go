package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/services"
	"server/utils"
)

type OrganizationController struct {
	orgService  *services.OrganizationService
	deptService *services.DepartmentService
}

func NewOrganizationController() *OrganizationController {
	return &OrganizationController{
		orgService:  &services.OrganizationService{},
		deptService: &services.DepartmentService{},
	}
}

func buildOrgTree(list []models.Organization) []gin.H {
	return buildFlatTree(list,
		func(item models.Organization) uint { return item.ID },
		func(item models.Organization) uint { return item.ParentID },
		func(item models.Organization) gin.H {
			return gin.H{
				"id":          item.ID,
				"parentId":    item.ParentID,
				"name":        item.Name,
				"code":        item.Code,
				"orgType":     item.OrgType,
				"orgTypeText": item.OrgTypeText(),
				"orgLevel":    item.OrgLevel,
				"category":    item.Category,
				"region":      item.Region,
				"province":    item.Province,
				"city":        item.City,
				"address":     item.Address,
				"manager":     item.Manager,
				"managerCode": item.ManagerCode,
				"sort":        item.Sort,
				"status":      item.Status,
				"description": item.Description,
				"userCount":   item.UserCount,
				"createdAt":   item.CreatedAt.Format("2006-01-02 15:04:05"),
				"children":    []gin.H{},
				"hasChildren": false,
				"departments": []gin.H{},
			}
		})
}

func (c *OrganizationController) GetOrganizations(ctx *gin.Context) {
	name := ctx.Query("name")
	orgTypeStr := ctx.Query("orgType")
	statusStr := ctx.Query("status")

	var orgType, status *int
	if orgTypeStr != "" {
		val, _ := strconv.Atoi(orgTypeStr)
		orgType = &val
	}
	if statusStr != "" {
		val, _ := strconv.Atoi(statusStr)
		status = &val
	}

	result, err := c.orgService.GetOrganizationList(name, orgType, status)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取机构列表失败"))
		return
	}

	tree := buildOrgTree(result.List)
	orgNameMap := buildOrgNameMap(result.List)
	for i := range tree {
		if err := c.attachOrgDepartments(tree[i], orgNameMap); err != nil {
			ctx.JSON(http.StatusOK, utils.Error(1, "获取机构部门失败"))
			return
		}
	}
	ctx.JSON(http.StatusOK, utils.Success("获取机构列表成功", utils.AllData(tree, result.Total)))
}

func (c *OrganizationController) attachOrgDepartments(node gin.H, orgNameMap map[uint]string) error {
	id, ok := node["id"].(uint)
	if !ok {
		return nil
	}
	depts, err := c.deptService.GetDepartmentsByOrgID(id)
	if err != nil {
		return err
	}
	deptTree := buildDeptTree(depts, orgNameMap)
	node["departments"] = deptTree
	if len(deptTree) > 0 {
		node["hasChildren"] = true
	}
	children, ok := node["children"].([]gin.H)
	if ok {
		for i := range children {
			if err := c.attachOrgDepartments(children[i], orgNameMap); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *OrganizationController) GetOrganizationTree(ctx *gin.Context) {
	list, err := c.orgService.GetOrganizationTree()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取机构树失败"))
		return
	}
	tree := buildOrgTree(list)
	ctx.JSON(http.StatusOK, utils.Success("获取机构树成功", tree))
}

func (c *OrganizationController) CreateOrganization(ctx *gin.Context) {
	var req models.Organization
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}
	// 忽略请求体携带的主键：避免指定 ID 写入
	req.ID = 0
	if err := c.orgService.CreateOrganization(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("创建机构失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("创建机构成功", nil))
}

func (c *OrganizationController) UpdateOrganization(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "机构ID无效"))
		return
	}

	var req models.Organization
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}

	if err := c.orgService.UpdateOrganization(uint(id), &req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("更新机构失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新机构成功", nil))
}

func (c *OrganizationController) DeleteOrganization(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "机构ID无效"))
		return
	}

	if err := c.orgService.DeleteOrganization(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("删除机构失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("删除机构成功", nil))
}

func (c *OrganizationController) GetOrganizationUsers(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "机构ID无效"))
		return
	}

	userIds, err := c.orgService.GetOrganizationUsers(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SafeErrText(err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取机构用户成功", userIds))
}

func (c *OrganizationController) AssignOrganizationUsers(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "机构ID无效"))
		return
	}

	var req struct {
		UserIds []int `json:"userIds"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误"))
		return
	}

	if err := c.orgService.AssignOrganizationUsers(uint(id), req.UserIds); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("分配用户失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("分配用户成功", nil))
}
