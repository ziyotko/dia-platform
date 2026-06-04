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
	userService *services.UserService
}

func NewDepartmentController() *DepartmentController {
	return &DepartmentController{
		deptService: &services.DepartmentService{},
		userService: &services.UserService{},
	}
}

func buildDeptTree(list []models.Department) []gin.H {
	nodeMap := make(map[uint]*gin.H)
	var roots []gin.H

	for i := range list {
		item := list[i]
		node := gin.H{
			"id":          item.ID,
			"parentId":    item.ParentID,
			"name":        item.Name,
			"code":        item.Code,
			"leader":      item.Leader,
			"sort":        item.Sort,
			"status":      item.Status,
			"description": item.Description,
			"createTime":  item.CreatedAt.Format("2006-01-02 15:04:05"),
			"children":    []gin.H{},
			"hasChildren": false,
		}
		nodeMap[item.ID] = &node
	}

	for i := range list {
		item := list[i]
		node := nodeMap[item.ID]
		if item.ParentID == 0 {
			roots = append(roots, *node)
		} else {
			if parent, ok := nodeMap[item.ParentID]; ok {
				children := (*parent)["children"].([]gin.H)
				children = append(children, *node)
				(*parent)["children"] = children
				(*parent)["hasChildren"] = true
			}
		}
	}

	return roots
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

	tree := buildDeptTree(result.List)
	ctx.JSON(http.StatusOK, utils.Success("获取部门列表成功", gin.H{
		"list":  tree,
		"total": result.Total,
	}))
}

func (c *DepartmentController) GetDepartmentTree(ctx *gin.Context) {
	list, err := c.deptService.GetDepartmentTree()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取部门树失败"))
		return
	}
	tree := buildDeptTree(list)
	ctx.JSON(http.StatusOK, utils.Success("获取部门树成功", tree))
}

func (c *DepartmentController) CreateDepartment(ctx *gin.Context) {
	var req models.Department
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: "+err.Error()))
		return
	}
	if err := c.deptService.CreateDepartment(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "创建部门失败: "+err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("创建部门成功", nil))
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
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: "+err.Error()))
		return
	}

	if err := c.deptService.UpdateDepartment(uint(id), &req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "更新部门失败: "+err.Error()))
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
		ctx.JSON(http.StatusOK, utils.Error(1, "删除部门失败: "+err.Error()))
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
		ctx.JSON(http.StatusOK, utils.Error(1, err.Error()))
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
		ctx.JSON(http.StatusOK, utils.Error(1, "分配用户失败: "+err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("分配用户成功", nil))
}
