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

type WorkflowController struct {
	workflowService *services.WorkflowService
}

func NewWorkflowController() *WorkflowController {
	return &WorkflowController{
		workflowService: &services.WorkflowService{},
	}
}

func (c *WorkflowController) GetWorkflows(ctx *gin.Context) {
	name := ctx.Query("name")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))

	// all=1 时不分页返回全量（供栏目绑定流程等下拉使用）
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

	workflows, total, err := c.workflowService.GetWorkflows(name, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取流程列表失败"))
		return
	}

	list := make([]gin.H, 0, len(workflows))
	for _, w := range workflows {
		nodeCount := len(w.Nodes)
		list = append(list, gin.H{
			"id":          w.ID,
			"name":        w.Name,
			"status":      w.Status,
			"description": w.Description,
			"nodeCount":   nodeCount,
			"createTime":  w.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	ctx.JSON(http.StatusOK, utils.Success("获取流程列表成功", gin.H{
		"list":     list,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}))
}

func (c *WorkflowController) GetWorkflowByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "流程ID无效"))
		return
	}

	workflow, err := c.workflowService.GetWorkflowByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取流程失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("获取流程成功", workflow))
}

func (c *WorkflowController) CreateWorkflow(ctx *gin.Context) {
	var req models.Workflow
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}

	if err := c.workflowService.CreateWorkflow(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("创建流程失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("创建流程成功", nil))
}

func (c *WorkflowController) UpdateWorkflow(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "流程ID无效"))
		return
	}

	var req models.Workflow
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}

	if err := c.workflowService.UpdateWorkflow(uint(id), &req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("更新流程失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新流程成功", nil))
}

func (c *WorkflowController) DeleteWorkflow(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "流程ID无效"))
		return
	}

	if err := c.workflowService.DeleteWorkflow(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("删除流程失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("删除流程成功", nil))
}

type SaveNodesReq struct {
	Nodes []models.WorkflowNode `json:"nodes"`
}

func (c *WorkflowController) SaveWorkflowNodes(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "流程ID无效"))
		return
	}

	var req SaveNodesReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}

	if err := c.workflowService.SaveWorkflowNodes(uint(id), req.Nodes); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("保存流程节点失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("保存流程节点成功", nil))
}

func (c *WorkflowController) GetWorkflowNodes(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "流程ID无效"))
		return
	}

	nodes, err := c.workflowService.GetWorkflowNodes(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取流程节点失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取流程节点成功", nodes))
}
