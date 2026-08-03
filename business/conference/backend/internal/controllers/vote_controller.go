package controllers

import (
	"strconv"

	"conference/internal/middleware"
	"conference/internal/models"
	"conference/internal/service"
	"conference/pkg/response"

	"github.com/gin-gonic/gin"
)

type VoteController struct {
	service service.VoteService
}

// --- Admin endpoints ---

func (ctrl *VoteController) Create(c *gin.Context) {
	var req struct {
		Vote    models.Vote         `json:"vote"`
		Options []models.VoteOption `json:"options"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.Create(&req.Vote, req.Options); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, req.Vote)
}

func (ctrl *VoteController) Update(c *gin.Context) {
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

func (ctrl *VoteController) Delete(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.service.Delete(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctrl *VoteController) GetByID(c *gin.Context) {
	id := parseUint(c.Param("id"))
	vote, err := ctrl.service.GetByID(id)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, vote)
}

func (ctrl *VoteController) List(c *gin.Context) {
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

func (ctrl *VoteController) GetResults(c *gin.Context) {
	id := parseUint(c.Param("id"))
	results, err := ctrl.service.GetResults(id)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, results)
}

// --- Member endpoints ---

func (ctrl *VoteController) ListAvailable(c *gin.Context) {
	userID := middleware.GetUserID(c)
	list, err := ctrl.service.ListAvailable(userID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, list)
}

func (ctrl *VoteController) CastVote(c *gin.Context) {
	voteID := parseUint(c.Param("id"))
	userID := middleware.GetUserID(c)
	var req struct {
		OptionID  uint64   `json:"optionId"`
		OptionIDs []uint64 `json:"optionIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.CastVote(voteID, userID, req.OptionID, req.OptionIDs); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "投票成功", nil)
}

func (ctrl *VoteController) HasVoted(c *gin.Context) {
	voteID := parseUint(c.Param("id"))
	userID := middleware.GetUserID(c)
	voted := ctrl.service.HasUserVoted(voteID, userID)
	response.Ok(c, map[string]bool{"voted": voted})
}
