package controllers

import (
	"strconv"

	"conference/internal/middleware"
	"conference/internal/service"
	"conference/pkg/response"

	"github.com/gin-gonic/gin"
)

type SignInController struct {
	service service.SignInService
}

// --- Member endpoints ---

func (ctrl *SignInController) GenerateQRCode(c *gin.Context) {
	meetingID := parseUint(c.Param("id"))
	userID := middleware.GetUserID(c)
	regID := parseUint(c.Query("registrationId"))

	token, err := ctrl.service.GenerateQRCodeToken(meetingID, userID, regID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, map[string]interface{}{"token": token})
}

func (ctrl *SignInController) GetStatus(c *gin.Context) {
	meetingID := parseUint(c.Param("id"))
	userID := middleware.GetUserID(c)

	status, err := ctrl.service.GetUserSignInStatus(meetingID, userID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, status)
}

func (ctrl *SignInController) SignInOnline(c *gin.Context) {
	meetingID := parseUint(c.Param("id"))
	userID := middleware.GetUserID(c)

	if err := ctrl.service.SignInOnline(meetingID, userID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "签到成功", nil)
}

func (ctrl *SignInController) SignOut(c *gin.Context) {
	meetingID := parseUint(c.Param("id"))
	userID := middleware.GetUserID(c)

	if err := ctrl.service.SignOut(meetingID, userID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "签退成功", nil)
}

// --- Admin endpoints ---

func (ctrl *SignInController) SignInByQR(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请提供签到码")
		return
	}
	if err := ctrl.service.SignInByQRCode(req.Token); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "签到成功", nil)
}

func (ctrl *SignInController) GetStats(c *gin.Context) {
	meetingID := parseUint(c.Param("id"))
	stats, err := ctrl.service.GetSignInStats(meetingID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, stats)
}

func (ctrl *SignInController) List(c *gin.Context) {
	meetingID := parseUint(c.Query("meetingId"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	list, total, err := ctrl.service.ListSignIns(meetingID, page, size)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}
