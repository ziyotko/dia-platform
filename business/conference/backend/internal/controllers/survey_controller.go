package controllers

import (
	"strconv"

	"conference/internal/middleware"
	"conference/internal/models"
	"conference/internal/service"
	"conference/pkg/response"

	"github.com/gin-gonic/gin"
)

type SurveyController struct {
	service service.SurveyService
}

// --- Admin endpoints ---

func (ctrl *SurveyController) Create(c *gin.Context) {
	var req struct {
		Survey    models.Survey           `json:"survey"`
		Questions []models.SurveyQuestion `json:"questions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.Create(&req.Survey, req.Questions); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, req.Survey)
}

func (ctrl *SurveyController) Update(c *gin.Context) {
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

func (ctrl *SurveyController) Delete(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.service.Delete(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctrl *SurveyController) GetByID(c *gin.Context) {
	id := parseUint(c.Param("id"))
	survey, err := ctrl.service.GetByID(id)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, survey)
}

func (ctrl *SurveyController) List(c *gin.Context) {
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

func (ctrl *SurveyController) GetResults(c *gin.Context) {
	id := parseUint(c.Param("id"))
	results, err := ctrl.service.GetResults(id)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, results)
}

func (ctrl *SurveyController) GetAnswerDetails(c *gin.Context) {
	id := parseUint(c.Param("id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	list, total, err := ctrl.service.GetAnswerDetails(id, page, size)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

// --- Member endpoints ---

func (ctrl *SurveyController) ListAvailable(c *gin.Context) {
	userID := middleware.GetUserID(c)
	list, err := ctrl.service.ListAvailable(userID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, list)
}

func (ctrl *SurveyController) SubmitAnswers(c *gin.Context) {
	surveyID := parseUint(c.Param("id"))
	userID := middleware.GetUserID(c)
	var req struct {
		Answers []struct {
			QuestionID uint64 `json:"questionId"`
			Answer     string `json:"answer"`
		} `json:"answers"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.SubmitAnswers(surveyID, userID, req.Answers); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "提交成功", nil)
}

func (ctrl *SurveyController) HasSubmitted(c *gin.Context) {
	surveyID := parseUint(c.Param("id"))
	userID := middleware.GetUserID(c)
	submitted := ctrl.service.HasUserSubmitted(surveyID, userID)
	response.Ok(c, map[string]bool{"submitted": submitted})
}
