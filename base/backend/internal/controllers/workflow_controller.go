package controllers

import (
	"strconv"

	"base/internal/models"
	"base/internal/service"
	"base/pkg/db"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

// WorkflowController 流程定义（含节点编排）管理。
type WorkflowController struct {
	service service.WorkflowService
}

func (ctl *WorkflowController) Create(c *gin.Context) {
	var wf models.Workflow
	if err := c.ShouldBindJSON(&wf); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	wf.TenantID = resolveTenantID(c, wf.TenantID)
	if wf.Code == "" {
		response.FailWithCode(c, response.CodeBadRequest, "请填写流程编码")
		return
	}
	if wf.Name == "" {
		response.FailWithCode(c, response.CodeBadRequest, "请填写流程名称")
		return
	}
	if err := ctl.service.Create(&wf); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, wf)
}

func (ctl *WorkflowController) Update(c *gin.Context) {
	var wf models.Workflow
	if err := c.ShouldBindJSON(&wf); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	wf.ID = uint64(parseID(c))
	tenantID := c.GetUint64("tenantID")
	// 租户不允许变更：取库中真实归属，避免请求体伪造 tenantId
	existing, err := ctl.service.GetByID(wf.ID, tenantID)
	if err != nil {
		response.Fail(c, "流程定义不存在")
		return
	}
	wf.TenantID = existing.TenantID
	if err := ctl.service.Update(&wf, tenantID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, wf)
}

func (ctl *WorkflowController) Delete(c *gin.Context) {
	if err := ctl.service.Delete(uint64(parseID(c)), c.GetUint64("tenantID")); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctl *WorkflowController) Get(c *gin.Context) {
	wf, err := ctl.service.GetByID(uint64(parseID(c)), c.GetUint64("tenantID"))
	if err != nil {
		response.Fail(c, "流程定义不存在")
		return
	}
	response.Ok(c, wf)
}

func (ctl *WorkflowController) List(c *gin.Context) {
	filterTenantID, _ := strconv.ParseUint(c.Query("tenantId"), 10, 64)
	page, size := pagination(c)
	list, total, err := ctl.service.List(service.WorkflowListQuery{
		TenantID:       c.GetUint64("tenantID"),
		FilterTenantID: filterTenantID,
		Keyword:        c.Query("keyword"),
		Status:         optionalIntQuery(c, "status"),
		Page:           page,
		Size:           size,
	})
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

// SaveNodes 覆盖式保存节点编排。
func (ctl *WorkflowController) SaveNodes(c *gin.Context) {
	var req struct {
		Nodes []models.WorkflowNode `json:"nodes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	if err := ctl.service.SaveNodes(uint64(parseID(c)), req.Nodes, c.GetUint64("tenantID")); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "节点保存成功", nil)
}

// Options 启用中的流程定义（发起流程时选择）。
func (ctl *WorkflowController) Options(c *gin.Context) {
	list, err := ctl.service.Options(c.GetUint64("tenantID"))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, list)
}

// ApproverOptions 节点审批人可选值（本租户的流程角色与用户）。
func (ctl *WorkflowController) ApproverOptions(c *gin.Context) {
	opts, err := ctl.service.ApproverOptions(c.GetUint64("tenantID"))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, opts)
}

// pagination 解析分页参数，统一上下限，避免超大 size 拖垮数据库。
func pagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 200 {
		size = 10
	}
	return page, size
}

// optionalIntQuery 解析可选的整型查询参数（空值返回 nil，表示不过滤）。
func optionalIntQuery(c *gin.Context, key string) *int {
	raw := c.Query(key)
	if raw == "" {
		return nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return nil
	}
	return &value
}

// workflowActor 组装流程引擎的操作者信息（含是否租户管理员）。
func workflowActor(c *gin.Context) service.WorkflowActor {
	return service.WorkflowActor{
		UserID:   c.GetUint64("userID"),
		Username: c.GetString("username"),
		TenantID: c.GetUint64("tenantID"),
		IsAdmin:  isAdminUser(c.GetUint64("userID")),
	}
}

// isAdminUser 判断用户是否为租户管理员（base_user.is_admin）。
func isAdminUser(userID uint64) bool {
	var count int64
	if err := db.DB.Model(&models.User{}).
		Where("id = ? AND is_admin = ?", userID, true).
		Count(&count).Error; err != nil {
		return false
	}
	return count > 0
}
