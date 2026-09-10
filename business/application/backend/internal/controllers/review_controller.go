package controllers

import (
	"application/internal/middleware"
	"application/internal/service"
	"application/pkg/response"

	"github.com/gin-gonic/gin"
)

type ReviewController struct {
	service service.ReviewService
}

// ListReviewers returns all reviewer accounts (admin use for assignment)
func (ctrl *ReviewController) ListReviewers(c *gin.Context) {
	list, err := ctrl.service.ListReviewers()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, list)
}

// MyAssignments lists the current reviewer's review tasks
func (ctrl *ReviewController) MyAssignments(c *gin.Context) {
	reviewerID := middleware.GetAdminID(c)
	page, size := getPage(c)
	status := c.Query("status")
	list, total, err := ctrl.service.MyAssignments(reviewerID, page, size, status)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

// GetAssignment returns detail of one review task
func (ctrl *ReviewController) GetAssignment(c *gin.Context) {
	reviewerID := middleware.GetAdminID(c)
	id := parseUint(c.Param("id"))
	assignment, err := ctrl.service.GetAssignment(id, reviewerID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, assignment)
}

type SubmitReviewReq struct {
	Score   float64 `json:"score"`
	Comment string  `json:"comment"`
}

func (ctrl *ReviewController) SubmitReview(c *gin.Context) {
	reviewerID := middleware.GetAdminID(c)
	id := parseUint(c.Param("id"))
	var req SubmitReviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	appSvc := service.ApplicationService{}
	if err := appSvc.SubmitReview(id, reviewerID, req.Score, req.Comment); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "专家评审", "评审打分", "assignment="+c.Param("id"))
	response.OkWithMessage(c, "评审提交成功", nil)
}

// ListAssignments lists assignments of an application (admin view)
func (ctrl *ReviewController) ListAssignments(c *gin.Context) {
	applicationID := parseUint(c.Query("applicationId"))
	list, err := ctrl.service.ListAssignments(applicationID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, list)
}
