package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/services"
	"server/utils"
)

type WorkflowRoleController struct {
	workflowRoleService *services.WorkflowRoleService
}

func NewWorkflowRoleController() *WorkflowRoleController {
	return &WorkflowRoleController{
		workflowRoleService: &services.WorkflowRoleService{},
	}
}

func (c *WorkflowRoleController) GetWorkflowRoles(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	name := ctx.Query("name")

	var status *int
	if statusStr := ctx.Query("status"); statusStr != "" {
		s, _ := strconv.Atoi(statusStr)
		status = &s
	}

	// all=1 时不分页返回全量（供审批人下拉使用）
	all := ctx.Query("all") == "1" || strings.EqualFold(ctx.Query("all"), "true")
	if all {
		page, pageSize = 1, 0
	} else {
		if page < 1 {
			page = 1
		}
		if pageSize < 1 {
			pageSize = 10
		}
	}

	result, err := c.workflowRoleService.GetWorkflowRoleList(page, pageSize, name, status)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取流程角色列表失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("获取流程角色列表成功", gin.H{
		"list":  result.List,
		"total": result.Total,
	}))
}

func (c *WorkflowRoleController) GetWorkflowRoleByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "流程角色ID无效"))
		return
	}

	role, err := c.workflowRoleService.GetWorkflowRoleByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SafeErrText(err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取流程角色成功", role))
}

func (c *WorkflowRoleController) CreateWorkflowRole(ctx *gin.Context) {
	var req models.WorkflowRole
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}
	if err := c.workflowRoleService.CreateWorkflowRole(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("创建流程角色失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("创建流程角色成功", nil))
}

func (c *WorkflowRoleController) UpdateWorkflowRole(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "流程角色ID无效"))
		return
	}

	var req models.WorkflowRole
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}

	if err := c.workflowRoleService.UpdateWorkflowRole(uint(id), &req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("更新流程角色失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新流程角色成功", nil))
}

func (c *WorkflowRoleController) DeleteWorkflowRole(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "流程角色ID无效"))
		return
	}

	if err := c.workflowRoleService.DeleteWorkflowRole(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("删除流程角色失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("删除流程角色成功", nil))
}

func (c *WorkflowRoleController) GetWorkflowRoleUsers(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "流程角色ID无效"))
		return
	}

	users, err := c.workflowRoleService.GetWorkflowRoleUsers(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("获取成员失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取成员成功", users))
}

func (c *WorkflowRoleController) UpdateWorkflowRoleUsers(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "流程角色ID无效"))
		return
	}

	var req struct {
		UserIDs []uint `json:"userIds"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}

	if err := c.workflowRoleService.UpdateWorkflowRoleUsers(uint(id), req.UserIDs); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("更新成员失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新成员成功", nil))
}
