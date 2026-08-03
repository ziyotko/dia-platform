package controllers

import (
	"strconv"
	"time"

	"conference/internal/middleware"
	"conference/internal/service"
	"conference/pkg/response"

	"github.com/gin-gonic/gin"
)

type LiveController struct {
	service service.LiveService
}

// --- Admin endpoints ---

func (ctrl *LiveController) StartLive(c *gin.Context) {
	meetingID := parseUint(c.Param("meetingId"))
	config, err := ctrl.service.StartLive(meetingID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, config)
}

func (ctrl *LiveController) StopLive(c *gin.Context) {
	meetingID := parseUint(c.Param("meetingId"))
	if err := ctrl.service.StopLive(meetingID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "直播已结束", nil)
}

func (ctrl *LiveController) GetLiveConfig(c *gin.Context) {
	meetingID := parseUint(c.Param("meetingId"))
	config, err := ctrl.service.GetLiveConfig(meetingID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, config)
}

func (ctrl *LiveController) UploadVod(c *gin.Context) {
	meetingID := parseUint(c.Param("meetingId"))
	var req struct {
		VideoURL     string `json:"videoUrl" binding:"required"`
		ReplayExpiry string `json:"replayExpiry"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	// Parse replay expiry if provided
	var expiry *time.Time
	if req.ReplayExpiry != "" {
		t, err := time.Parse("2006-01-02 15:04:05", req.ReplayExpiry)
		if err == nil {
			expiry = &t
		}
	}
	if err := ctrl.service.UploadVod(meetingID, req.VideoURL, expiry); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "录播上传成功", nil)
}

func (ctrl *LiveController) SetReplayStatus(c *gin.Context) {
	meetingID := parseUint(c.Param("meetingId"))
	var req struct {
		AllowReplay bool `json:"allowReplay"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.SetReplayStatus(meetingID, req.AllowReplay); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "设置成功", nil)
}

func (ctrl *LiveController) GetViewingLogs(c *gin.Context) {
	meetingID := parseUint(c.Query("meetingId"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	list, total, err := ctrl.service.GetViewingLogs(meetingID, page, size)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

// --- Member endpoints ---

func (ctrl *LiveController) GetPlayURL(c *gin.Context) {
	meetingID := parseUint(c.Param("meetingId"))
	userID := middleware.GetUserID(c)

	url, err := ctrl.service.GetPlayURL(meetingID, userID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, map[string]string{"playUrl": url})
}

func (ctrl *LiveController) SendMessage(c *gin.Context) {
	meetingID := parseUint(c.Param("meetingId"))
	userID := middleware.GetUserID(c)
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "消息不能为空")
		return
	}
	if err := ctrl.service.SendLiveMessage(meetingID, userID, req.Content); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "发送成功", nil)
}

func (ctrl *LiveController) GetMessages(c *gin.Context) {
	meetingID := parseUint(c.Param("meetingId"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	msgs, err := ctrl.service.GetLiveMessages(meetingID, limit)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, msgs)
}

func (ctrl *LiveController) StartViewing(c *gin.Context) {
	meetingID := parseUint(c.Param("meetingId"))
	userID := middleware.GetUserID(c)
	viewType := c.DefaultQuery("type", "live")

	if err := ctrl.service.RecordViewingStart(meetingID, userID, viewType); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "记录已开始", nil)
}
