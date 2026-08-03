package controllers

import (
	"strconv"

	"conference/internal/middleware"
	"conference/internal/models"
	"conference/internal/service"
	"conference/pkg/response"

	"github.com/gin-gonic/gin"
)

type MeetingController struct {
	service service.MeetingService
}

// --- Admin endpoints ---

func (ctrl *MeetingController) Create(c *gin.Context) {
	var meeting models.Meeting
	if err := c.ShouldBindJSON(&meeting); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.Create(&meeting); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, meeting)
}

func (ctrl *MeetingController) Update(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.Update(id, updates); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "更新成功", nil)
}

func (ctrl *MeetingController) Delete(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.service.Delete(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctrl *MeetingController) GetByID(c *gin.Context) {
	id := parseUint(c.Param("id"))
	meeting, err := ctrl.service.GetByID(id)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, meeting)
}

func (ctrl *MeetingController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")
	meetingType := c.Query("type")
	status := c.Query("status")

	list, total, err := ctrl.service.List(page, size, keyword, meetingType, status, "")
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctrl *MeetingController) Close(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.service.CloseMeeting(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "会议已关闭", nil)
}

func (ctrl *MeetingController) SaveAgendas(c *gin.Context) {
	meetingID := parseUint(c.Param("id"))
	var agendas []models.Agenda
	if err := c.ShouldBindJSON(&agendas); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.SaveAgendas(meetingID, agendas); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "保存成功", nil)
}

func (ctrl *MeetingController) SaveGuests(c *gin.Context) {
	meetingID := parseUint(c.Param("id"))
	var guests []models.Guest
	if err := c.ShouldBindJSON(&guests); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.SaveGuests(meetingID, guests); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "保存成功", nil)
}

// --- Member endpoints ---

func (ctrl *MeetingController) ListAvailable(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")

	branch := middleware.GetBranch(c)
	level := middleware.GetMemberLevel(c)

	list, total, err := ctrl.service.ListAvailable(page, size, keyword, branch, level)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctrl *MeetingController) GetDetail(c *gin.Context) {
	id := parseUint(c.Param("id"))
	meeting, err := ctrl.service.GetByID(id)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, meeting)
}

func parseUint(s string) uint64 {
	v, _ := strconv.ParseUint(s, 10, 64)
	return v
}
