package controllers

import (
	"application/internal/middleware"
	"application/internal/models"
	"application/internal/service"
	"application/pkg/response"

	"github.com/gin-gonic/gin"
)

type ApplicationController struct {
	service service.ApplicationService
}

// --- Applicant endpoints (申报人) ---

func (ctrl *ApplicationController) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	// 只接收申报人可填写的字段。直接绑定 models.Application 会让请求体注入
	// status / published_at / submitted_at / total_score / final_opinion / id 等内部字段，
	// 从而把伪造条目直接塞进「结果公示」（该公示列表是全体申报人可见的）。
	// 与 UpdateDraft 的白名单思路保持一致。
	var req struct {
		BatchID      uint64 `json:"batchId"`
		CategoryID   uint64 `json:"categoryId"`
		Title        string `json:"title"`
		ProjectBrief string `json:"projectBrief"`
		Content      string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	app := models.Application{
		BatchID:      req.BatchID,
		CategoryID:   req.CategoryID,
		Title:        req.Title,
		ProjectBrief: req.ProjectBrief,
		Content:      req.Content,
		UserID:       userID,
	}
	if err := ctrl.service.Create(&app); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, app)
}

func (ctrl *ApplicationController) UpdateDraft(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := parseUint(c.Param("id"))
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.UpdateDraft(id, userID, updates); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "保存成功", nil)
}

func (ctrl *ApplicationController) DeleteDraft(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := parseUint(c.Param("id"))
	if err := ctrl.service.DeleteDraft(id, userID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctrl *ApplicationController) Submit(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := parseUint(c.Param("id"))
	if err := ctrl.service.Submit(id, userID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "提交成功，等待初审", nil)
}

// Withdraw takes back a submitted application before the preliminary review.
func (ctrl *ApplicationController) Withdraw(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := parseUint(c.Param("id"))
	if err := ctrl.service.Withdraw(id, userID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "已撤回，可继续修改", nil)
}

func (ctrl *ApplicationController) MyApplications(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, size := getPage(c)
	status := c.Query("status")
	keyword := c.Query("keyword")
	list, total, err := ctrl.service.ListUser(userID, page, size, status, keyword)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctrl *ApplicationController) GetMine(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := parseUint(c.Param("id"))
	app, err := ctrl.service.GetForUser(id, userID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, app)
}

func (ctrl *ApplicationController) SaveMaterials(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := parseUint(c.Param("id"))
	var materials []models.ApplicationMaterial
	if err := c.ShouldBindJSON(&materials); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.SaveMaterials(id, userID, materials); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "材料已保存", nil)
}

// --- Admin endpoints (管理人) ---

func (ctrl *ApplicationController) List(c *gin.Context) {
	page, size := getPage(c)
	batchID := parseUint(c.Query("batchId"))
	status := c.Query("status")
	statuses := c.Query("statuses")
	keyword := c.Query("keyword")
	list, total, err := ctrl.service.List(page, size, batchID, status, statuses, keyword)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctrl *ApplicationController) GetByID(c *gin.Context) {
	id := parseUint(c.Param("id"))
	app, err := ctrl.service.GetByID(id)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, app)
}

type PreliminaryReq struct {
	Pass    bool   `json:"pass"`
	Opinion string `json:"opinion"`
}

func (ctrl *ApplicationController) PreliminaryReview(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req PreliminaryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.PreliminaryReview(id, req.Pass, req.Opinion); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "初审管理", "在线初审", "id="+c.Param("id"))
	response.OkWithMessage(c, "初审完成", nil)
}

// RevokePreliminary 撤回初审：把误通过的申报退回「待初审」重新处理。
func (ctrl *ApplicationController) RevokePreliminary(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.service.RevokePreliminary(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "初审管理", "撤回初审", "id="+c.Param("id"))
	response.OkWithMessage(c, "已撤回初审，该申报回到「待初审」", nil)
}

type AssignReq struct {
	ReviewerIDs []uint64 `json:"reviewerIds"`
}

func (ctrl *ApplicationController) AssignReviewers(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req AssignReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.AssignReviewers(id, req.ReviewerIDs); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "评审管理", "分配评审专家", "id="+c.Param("id"))
	response.OkWithMessage(c, "评审分配成功", nil)
}

type FinalizeReq struct {
	Pass    bool   `json:"pass"`
	Opinion string `json:"opinion"`
}

func (ctrl *ApplicationController) Finalize(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req FinalizeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.Finalize(id, req.Pass, req.Opinion); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "评审管理", "确定评审结果", "id="+c.Param("id"))
	response.OkWithMessage(c, "评审结果已确定", nil)
}

func (ctrl *ApplicationController) PublishResult(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.service.PublishResult(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "结果公示", "公示申报结果", "id="+c.Param("id"))
	response.OkWithMessage(c, "结果已公示", nil)
}

// RevokeResult takes back a published/decided result so it can be corrected.
func (ctrl *ApplicationController) RevokeResult(c *gin.Context) {
	id := parseUint(c.Param("id"))
	message, err := ctrl.service.RevokeResult(id)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "结果公示", "撤回评审结果", "id="+c.Param("id"))
	response.OkWithMessage(c, message, nil)
}
