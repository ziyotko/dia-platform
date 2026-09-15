package controllers

import (
	"strconv"

	"base/internal/models"
	"base/internal/service"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

type MessageController struct {
	service service.MessageService
}

func (ctl *MessageController) Create(c *gin.Context) {
	var m models.Message
	if err := c.ShouldBindJSON(&m); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	m.SenderID = c.GetUint64("userID")
	m.SenderName = c.GetString("username")
	m.TenantID = c.GetUint64("tenantID")
	if err := ctl.service.Create(&m); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, m)
}

func (ctl *MessageController) Send(c *gin.Context) {
	var req struct {
		ReceiverIDs  []uint64 `json:"receiverIds"`
		ReceiverType string   `json:"receiverType"`
		Title        string   `json:"title" binding:"required"`
		Content      string   `json:"content" binding:"required"`
		Type         string   `json:"type"`
		Priority     string   `json:"priority"`
		Channel      string   `json:"channel"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	senderID := c.GetUint64("userID")
	senderName := c.GetString("username")
	tenantID := c.GetUint64("tenantID")

	if req.Type == "" {
		req.Type = "system"
	}
	if req.Priority == "" {
		req.Priority = "normal"
	}
	if req.Channel == "" {
		req.Channel = "in-app"
	}

	var err error
	if req.ReceiverType == "all" {
		err = ctl.service.Broadcast(senderID, senderName, tenantID, req.Title, req.Content, req.Type, req.Priority)
	} else {
		err = ctl.service.SendToUsers(senderID, senderName, tenantID, req.ReceiverIDs, req.Title, req.Content, req.Type, req.Priority)
	}
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	// 站外渠道发送（目前实现邮件）
	if req.Channel == "email" && req.ReceiverType != "all" {
		go ctl.sendEmailToUsers(req.ReceiverIDs, req.Title, req.Content)
	}

	response.OkWithMessage(c, "发送成功", nil)
}

func (ctl *MessageController) sendEmailToUsers(userIDs []uint64, subject, content string) {
	userSvc := service.UserService{}
	for _, uid := range userIDs {
		user, err := userSvc.GetByID(uid, 0)
		if err != nil || user.Email == "" {
			continue
		}
		_ = ctl.service.SendExternal("email", user.Email, subject, content)
	}
}

func (ctl *MessageController) Delete(c *gin.Context) {
	id := uint64(parseID(c))
	tenantID := c.GetUint64("tenantID")
	if err := ctl.service.Delete(id, tenantID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctl *MessageController) Get(c *gin.Context) {
	id := uint64(parseID(c))
	tenantID := c.GetUint64("tenantID")
	m, err := ctl.service.GetByID(id, tenantID)
	if err != nil {
		response.Fail(c, "消息不存在")
		return
	}
	response.Ok(c, m)
}

func (ctl *MessageController) List(c *gin.Context) {
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	isRead, _ := strconv.Atoi(c.DefaultQuery("isRead", "-1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	list, total, err := ctl.service.List(service.MessageListQuery{
		TenantID: c.GetUint64("tenantID"),
		UserID:   c.GetUint64("userID"),
		Box:      c.DefaultQuery("box", "inbox"),
		Status:   status,
		IsRead:   isRead,
		Page:     page,
		Size:     size,
	})
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctl *MessageController) UnreadCount(c *gin.Context) {
	receiverID := c.GetUint64("userID")
	count, err := ctl.service.GetUnreadCount(receiverID, c.GetUint64("tenantID"))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, count)
}

func (ctl *MessageController) MarkRead(c *gin.Context) {
	id := uint64(parseID(c))
	receiverID := c.GetUint64("userID")
	if err := ctl.service.MarkRead(id, receiverID, c.GetUint64("tenantID")); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "已标记为已读", nil)
}

// MarkAllRead 将当前用户的全部未读消息标为已读
func (ctl *MessageController) MarkAllRead(c *gin.Context) {
	count, err := ctl.service.MarkAllRead(c.GetUint64("userID"), c.GetUint64("tenantID"))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "已全部标为已读", gin.H{"count": count})
}
