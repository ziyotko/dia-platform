package controllers

import (
	"strconv"

	"conference/internal/middleware"
	"conference/internal/service"
	"conference/pkg/response"

	"github.com/gin-gonic/gin"
)

type NotificationController struct {
	service service.NotificationService
}

// --- Admin endpoints ---

func (ctrl *NotificationController) Send(c *gin.Context) {
	adminID := middleware.GetAdminID(c)
	var req struct {
		Title      string   `json:"title" binding:"required"`
		Content    string   `json:"content" binding:"required"`
		TargetType string   `json:"targetType" binding:"required"`
		TargetIDs  []uint64 `json:"targetIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.Send(adminID, req.Title, req.Content, req.TargetType, req.TargetIDs); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "发送成功", nil)
}

func (ctrl *NotificationController) ListSent(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	list, total, err := ctrl.service.ListSent(page, size)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctrl *NotificationController) GetReadReceipt(c *gin.Context) {
	id := parseUint(c.Param("id"))
	receipt, err := ctrl.service.GetReadReceipt(id)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, receipt)
}

// --- Member endpoints ---

func (ctrl *NotificationController) MyNotifications(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	list, total, err := ctrl.service.GetUserNotifications(userID, page, size)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctrl *NotificationController) MarkRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := parseUint(c.Param("id"))

	if err := ctrl.service.MarkRead(id, userID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "已标记已读", nil)
}

func (ctrl *NotificationController) GetUnreadCount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	count, err := ctrl.service.GetUnreadCount(userID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, map[string]int64{"count": count})
}
