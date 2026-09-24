package controllers

import (
	"strconv"

	"base/internal/models"
	"base/internal/service"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

// WorkflowEngineController 流程实例与审批任务（待办）接口。
//
// 这些接口在路由层被列入权限白名单：它们都是「自己的数据」——
// 发起、撤销只能由发起人（或管理员）操作，审批只能由该任务的审批人操作，
// 详情只能由发起人/参与者/管理员查看，归属校验在 service 层完成。
type WorkflowEngineController struct {
	service service.WorkflowEngineService
}

// Start 发起流程
func (ctl *WorkflowEngineController) Start(c *gin.Context) {
	var req service.StartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	if req.WorkflowID == 0 {
		response.FailWithCode(c, response.CodeBadRequest, "请选择流程")
		return
	}
	instance, err := ctl.service.Start(req, workflowActor(c))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, instance)
}

// ListInstances 流程实例列表。非管理员只能看到自己发起的实例。
func (ctl *WorkflowEngineController) ListInstances(c *gin.Context) {
	actor := workflowActor(c)
	page, size := pagination(c)
	filterTenantID, _ := strconv.ParseUint(c.Query("tenantId"), 10, 64)
	workflowID, _ := strconv.ParseUint(c.Query("workflowId"), 10, 64)

	mine := c.Query("mine") == "1"
	if !actor.IsAdmin && !models.IsPlatformTenant(actor.TenantID) {
		mine = true
	}

	list, total, err := ctl.service.ListInstances(service.WorkflowInstanceQuery{
		TenantID:       actor.TenantID,
		FilterTenantID: filterTenantID,
		Mine:           mine,
		UserID:         actor.UserID,
		WorkflowID:     workflowID,
		Status:         optionalIntQuery(c, "status"),
		Keyword:        c.Query("keyword"),
		Page:           page,
		Size:           size,
	})
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

// Detail 流程实例详情（含任务与流转日志）
func (ctl *WorkflowEngineController) Detail(c *gin.Context) {
	detail, err := ctl.service.GetInstance(uint64(parseID(c)), workflowActor(c))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, detail)
}

// Cancel 撤销流程（仅发起人或管理员）
func (ctl *WorkflowEngineController) Cancel(c *gin.Context) {
	if err := ctl.service.Cancel(uint64(parseID(c)), workflowActor(c)); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "已撤销", nil)
}

// Delete 删除流程实例（仅管理员）
func (ctl *WorkflowEngineController) Delete(c *gin.Context) {
	if err := ctl.service.Delete(uint64(parseID(c)), workflowActor(c)); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

// MyTasks 我的审批任务（box: todo 待办 / done 已办）
func (ctl *WorkflowEngineController) MyTasks(c *gin.Context) {
	actor := workflowActor(c)
	page, size := pagination(c)
	box := c.DefaultQuery("box", "todo")
	list, total, err := ctl.service.MyTasks(actor, box, page, size)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

// Approve 审批通过
func (ctl *WorkflowEngineController) Approve(c *gin.Context) {
	ctl.handleTaskAction(c, true)
}

// Reject 审批驳回
func (ctl *WorkflowEngineController) Reject(c *gin.Context) {
	ctl.handleTaskAction(c, false)
}

// ApproverOptions 转办/加签可选的用户列表
func (ctl *WorkflowEngineController) ApproverOptions(c *gin.Context) {
	options, err := ctl.service.ApproverOptions(workflowActor(c))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, options)
}

// Transfer 转办：把待办交给同租户的另一个用户（仅当前审批人可操作）
func (ctl *WorkflowEngineController) Transfer(c *gin.Context) {
	var req service.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	if err := ctl.service.Transfer(uint64(parseID(c)), req, workflowActor(c)); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "已转办", nil)
}

// AddApprover 加签：在当前节点追加一个审批人
func (ctl *WorkflowEngineController) AddApprover(c *gin.Context) {
	var req service.AddApproverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	if err := ctl.service.AddApprover(uint64(parseID(c)), req, workflowActor(c)); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "已加签", nil)
}

func (ctl *WorkflowEngineController) handleTaskAction(c *gin.Context, approve bool) {
	var req struct {
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	taskID := uint64(parseID(c))
	actor := workflowActor(c)
	// 审批意见按字符数截断（受限列长度约束，避免超长报 500）
	comment := clipComment(req.Comment)

	var err error
	message := "已驳回"
	if approve {
		err = ctl.service.Approve(taskID, comment, actor)
		message = "审批通过"
	} else {
		err = ctl.service.Reject(taskID, comment, actor)
	}
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, message, nil)
}

// clipComment 审批意见超长截断（WorkflowTask.Comment / WorkflowLog.Comment 为 512 字节级定长列）
func clipComment(s string) string {
	r := []rune(s)
	if len(r) > 500 {
		return string(r[:500])
	}
	return s
}
