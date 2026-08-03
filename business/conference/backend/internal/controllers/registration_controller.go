package controllers

import (
	"strconv"

	"conference/internal/middleware"
	"conference/internal/service"
	"conference/pkg/response"

	"github.com/gin-gonic/gin"
)

type RegistrationController struct {
	service service.RegistrationService
}

// --- Member endpoints ---

func (ctrl *RegistrationController) Register(c *gin.Context) {
	meetingID := parseUint(c.Param("meetingId"))
	userID := middleware.GetUserID(c)

	reg, err := ctrl.service.Register(meetingID, userID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, reg)
}

func (ctrl *RegistrationController) Cancel(c *gin.Context) {
	regID := parseUint(c.Param("id"))
	userID := middleware.GetUserID(c)

	if err := ctrl.service.Cancel(regID, userID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "取消报名成功", nil)
}

func (ctrl *RegistrationController) MyRegistrations(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	status := c.Query("status")

	list, total, err := ctrl.service.ListByUser(userID, status, page, size)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctrl *RegistrationController) MyRegistrationStatus(c *gin.Context) {
	meetingID := parseUint(c.Param("meetingId"))
	userID := middleware.GetUserID(c)

	reg, err := ctrl.service.GetMyRegistration(meetingID, userID)
	if err != nil {
		response.Ok(c, map[string]interface{}{"registered": false})
		return
	}
	response.Ok(c, reg)
}

// --- Admin endpoints ---

func (ctrl *RegistrationController) List(c *gin.Context) {
	meetingID := parseUint(c.Query("meetingId"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	status := c.Query("status")

	list, total, err := ctrl.service.List(meetingID, status, page, size)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctrl *RegistrationController) Approve(c *gin.Context) {
	id := parseUint(c.Param("id"))
	adminID := middleware.GetAdminID(c)
	var req struct {
		Comment string `json:"comment"`
	}
	c.ShouldBindJSON(&req)

	if err := ctrl.service.Approve(id, adminID, req.Comment); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "审核通过", nil)
}

func (ctrl *RegistrationController) Reject(c *gin.Context) {
	id := parseUint(c.Param("id"))
	adminID := middleware.GetAdminID(c)
	var req struct {
		Comment string `json:"comment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写驳回原因")
		return
	}
	if err := ctrl.service.Reject(id, adminID, req.Comment); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "已驳回", nil)
}

func (ctrl *RegistrationController) PromoteWaitlist(c *gin.Context) {
	id := parseUint(c.Param("id"))
	adminID := middleware.GetAdminID(c)

	if err := ctrl.service.PromoteFromWaitlist(id, adminID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "已转为正式报名", nil)
}

func (ctrl *RegistrationController) GetWaitlist(c *gin.Context) {
	meetingID := parseUint(c.Query("meetingId"))
	list, err := ctrl.service.GetWaitlist(meetingID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, list)
}
