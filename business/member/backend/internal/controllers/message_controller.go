package controllers

import (
	"member/internal/middleware"
	"member/internal/service"
	"member/pkg/response"

	"github.com/gin-gonic/gin"
)

type MessageController struct {
	msgService service.MessageService
}

func (ctrl *MessageController) CreateMessage(c *gin.Context) {
	memberID := middleware.GetMemberID(c)
	var req service.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整信息")
		return
	}
	msg, err := ctrl.msgService.CreateMessage(memberID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "留言成功", msg)
}

func (ctrl *MessageController) GetMyMessages(c *gin.Context) {
	memberID := middleware.GetMemberID(c)
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 10)

	msgs, total, err := ctrl.msgService.GetMyMessages(memberID, page, size)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"list":  msgs,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

func (ctrl *MessageController) GetMessage(c *gin.Context) {
	id := parseUint(c.Param("id"))
	msg, err := ctrl.msgService.GetMessage(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, msg)
}

func (ctrl *MessageController) ListAllMessages(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 10)
	status := c.Query("status")

	msgs, total, err := ctrl.msgService.ListAllMessages(page, size, status)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"list":  msgs,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

func (ctrl *MessageController) ReplyMessage(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req struct {
		Reply string `json:"reply" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写回复内容")
		return
	}
	if err := ctrl.msgService.ReplyMessage(id, req.Reply); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "回复成功", nil)
}

func (ctrl *MessageController) DeleteMessage(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.msgService.DeleteMessage(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}
