package controllers

import (
	"member/internal/middleware"
	"member/internal/service"
	"member/pkg/response"

	"github.com/gin-gonic/gin"
)

type ApplicationController struct {
	appService service.ApplicationService
}

// CreateApplication submits a new application
func (ctrl *ApplicationController) CreateApplication(c *gin.Context) {
	memberID := middleware.GetMemberID(c)
	var req service.CreateAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整信息")
		return
	}

	app, err := ctrl.appService.CreateApplication(memberID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "申请提交成功", app)
}

// WithdrawApplication withdraws a pending application
func (ctrl *ApplicationController) WithdrawApplication(c *gin.Context) {
	memberID := middleware.GetMemberID(c)
	id := parseUint(c.Param("id"))

	if err := ctrl.appService.WithdrawApplication(id, memberID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "申请已撤回", nil)
}

// GetMyApplications returns member's applications
func (ctrl *ApplicationController) GetMyApplications(c *gin.Context) {
	memberID := middleware.GetMemberID(c)
	apps, err := ctrl.appService.GetMyApplications(memberID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, apps)
}

// GetApplication returns an application detail (owner only)
func (ctrl *ApplicationController) GetApplication(c *gin.Context) {
	memberID := middleware.GetMemberID(c)
	id := parseUint(c.Param("id"))
	app, err := ctrl.appService.GetApplication(id, memberID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, app)
}

// ReviewApplication reviews an application (admin)
func (ctrl *ApplicationController) ReviewApplication(c *gin.Context) {
	id := parseUint(c.Param("id"))
	reviewerID := middleware.GetMemberID(c)

	var req struct {
		Approved bool   `json:"approved"`
		Comment  string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := ctrl.appService.ReviewApplication(id, reviewerID, req.Approved, req.Comment); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "审核完成", nil)
}

// ListApplications lists all applications (admin)
func (ctrl *ApplicationController) ListApplications(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 10)
	status := c.Query("status")

	apps, total, err := ctrl.appService.ListApplications(page, size, status)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"list":  apps,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

func parseUint(s string) uint64 {
	var n uint64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + uint64(c-'0')
		}
	}
	return n
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	if n == 0 {
		return def
	}
	return n
}
