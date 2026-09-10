package controllers

import (
	"application/internal/middleware"
	"application/internal/service"
	"application/pkg/response"

	"github.com/gin-gonic/gin"
)

type NotificationController struct {
	service service.NotificationService
}

// MyNotifications lists the current applicant's notifications (frontend)
func (ctrl *NotificationController) MyNotifications(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, size := getPage(c)
	list, total, err := ctrl.service.ListUser(userID, page, size)
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
	response.OkWithMessage(c, "已读", nil)
}

func (ctrl *NotificationController) UnreadCount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	count := ctrl.service.UnreadCount(userID)
	response.Ok(c, gin.H{"count": count})
}

// --- Admin ---

type SendReq struct {
	UserID  uint64 `json:"userId"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Type    string `json:"type"`
}

func (ctrl *NotificationController) Send(c *gin.Context) {
	var req SendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.Send(req.Title, req.Content, req.Type, req.UserID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "通知管理", "发送通知", req.Title)
	response.OkWithMessage(c, "通知已发送", nil)
}
