package controllers

import (
	"strconv"

	"base/internal/models"
	"base/internal/service"
	"base/pkg/response"
	"base/pkg/utils"

	"github.com/gin-gonic/gin"
)

// WorkflowRoleController 流程角色（审批角色）管理。
type WorkflowRoleController struct {
	service service.WorkflowRoleService
}

func (ctl *WorkflowRoleController) Create(c *gin.Context) {
	var r models.WorkflowRole
	if err := c.ShouldBindJSON(&r); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	r.TenantID = resolveTenantID(c, r.TenantID)
	if r.Code == "" {
		r.Code = "wfrole_" + utils.RandomDigit(6)
	}
	if err := ctl.service.Create(&r); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, r)
}

func (ctl *WorkflowRoleController) Update(c *gin.Context) {
	var r models.WorkflowRole
	if err := c.ShouldBindJSON(&r); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	r.ID = uint64(parseID(c))
	// 租户不允许变更，取数据库中的归属，避免请求体伪造 tenantId 绕过校验
	tenantID := c.GetUint64("tenantID")
	existing, err := ctl.service.GetByID(r.ID, tenantID)
	if err != nil {
		response.Fail(c, "流程角色不存在")
		return
	}
	r.TenantID = existing.TenantID
	if err := ctl.service.Update(&r, tenantID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, r)
}

func (ctl *WorkflowRoleController) Delete(c *gin.Context) {
	id := uint64(parseID(c))
	tenantID := c.GetUint64("tenantID")
	if err := ctl.service.Delete(id, tenantID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctl *WorkflowRoleController) Get(c *gin.Context) {
	id := uint64(parseID(c))
	tenantID := c.GetUint64("tenantID")
	r, err := ctl.service.GetByID(id, tenantID)
	if err != nil {
		response.Fail(c, "流程角色不存在")
		return
	}
	response.Ok(c, r)
}

func (ctl *WorkflowRoleController) List(c *gin.Context) {
	tenantID := c.GetUint64("tenantID")
	filterTenantID, _ := strconv.ParseUint(c.Query("tenantId"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 200 {
		size = 10
	}
	keyword := c.Query("keyword")

	var status *int
	if v := c.Query("status"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			status = &parsed
		}
	}

	list, total, err := ctl.service.List(tenantID, filterTenantID, page, size, keyword, status)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

// AssignUsers 覆盖式设置流程角色成员。
func (ctl *WorkflowRoleController) AssignUsers(c *gin.Context) {
	id := uint64(parseID(c))
	var req struct {
		UserIDs []uint64 `json:"userIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	if err := ctl.service.AssignUsers(id, req.UserIDs, c.GetUint64("tenantID")); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "成员保存成功", nil)
}

// UserOptions 成员选择器用的用户选项。
// 单独提供该接口，是为了让只被授予「流程角色」的管理员也能配置成员，
// 而不必额外拥有用户管理（base:user:list）权限。
func (ctl *WorkflowRoleController) UserOptions(c *gin.Context) {
	tenantID := c.GetUint64("tenantID")
	filterTenantID, _ := strconv.ParseUint(c.Query("tenantId"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	users, err := ctl.service.ListUserOptions(tenantID, filterTenantID, c.Query("keyword"), limit)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	// 只返回选择器需要的字段，避免泄露手机号/邮箱等
	type option struct {
		ID       uint64 `json:"id"`
		TenantID uint64 `json:"tenantId"`
		Username string `json:"username"`
		RealName string `json:"realName"`
		Status   int    `json:"status"`
	}
	options := make([]option, 0, len(users))
	for _, u := range users {
		options = append(options, option{
			ID:       u.ID,
			TenantID: u.TenantID,
			Username: u.Username,
			RealName: u.RealName,
			Status:   u.Status,
		})
	}
	response.Ok(c, options)
}
