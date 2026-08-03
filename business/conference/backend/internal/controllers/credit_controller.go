package controllers

import (
	"strconv"

	"conference/internal/middleware"
	"conference/internal/service"
	"conference/pkg/response"

	"github.com/gin-gonic/gin"
)

type CreditController struct {
	service service.CreditService
}

// --- Admin endpoints ---

func (ctrl *CreditController) SetMeetingCredits(c *gin.Context) {
	meetingID := parseUint(c.Param("id"))
	var req struct {
		Credits float64 `json:"credits"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.SetMeetingCredits(meetingID, req.Credits); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "设置成功", nil)
}

func (ctrl *CreditController) ManualAdjust(c *gin.Context) {
	var req struct {
		UserID    uint64  `json:"userId" binding:"required"`
		MeetingID uint64  `json:"meetingId" binding:"required"`
		Credits   float64 `json:"credits" binding:"required"`
		Remark    string  `json:"remark" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	operatorID := middleware.GetAdminID(c)
	if err := ctrl.service.ManualAdjust(req.UserID, req.MeetingID, operatorID, req.Credits, req.Remark); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "操作成功", nil)
}

func (ctrl *CreditController) ListRecords(c *gin.Context) {
	meetingID := parseUint(c.Query("meetingId"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	source := c.Query("source")

	list, total, err := ctrl.service.ListRecords(meetingID, source, page, size)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctrl *CreditController) GetStats(c *gin.Context) {
	stats, err := ctrl.service.GetStats()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, stats)
}

// --- Member endpoints ---

func (ctrl *CreditController) MyCredits(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	total, list, count, err := ctrl.service.GetUserCredits(userID, page, size)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, map[string]interface{}{
		"total_credits": total,
		"records":       list,
		"record_count":  count,
	})
}
