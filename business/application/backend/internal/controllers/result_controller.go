package controllers

import (
	"application/internal/middleware"
	"application/internal/models"
	"application/internal/service"
	"application/pkg/response"

	"github.com/gin-gonic/gin"
)

type ResultController struct {
	service service.ResultService
}

// --- Announcements (结果公示) ---

func (ctrl *ResultController) CreateAnnouncement(c *gin.Context) {
	var a models.Announcement
	if err := c.ShouldBindJSON(&a); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.CreateAnnouncement(&a); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "结果公示", "新增公示", a.Title)
	response.Ok(c, a)
}

func (ctrl *ResultController) UpdateAnnouncement(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.UpdateAnnouncement(id, updates); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "结果公示", "编辑公示", "id="+c.Param("id"))
	response.OkWithMessage(c, "更新成功", nil)
}

func (ctrl *ResultController) DeleteAnnouncement(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.service.DeleteAnnouncement(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "结果公示", "删除公示", "id="+c.Param("id"))
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctrl *ResultController) PublishAnnouncement(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.service.PublishAnnouncement(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "结果公示", "发布公示", "id="+c.Param("id"))
	response.OkWithMessage(c, "公示已发布", nil)
}

func (ctrl *ResultController) ListAnnouncements(c *gin.Context) {
	page, size := getPage(c)
	keyword := c.Query("keyword")
	list, total, err := ctrl.service.ListAnnouncements(page, size, keyword, false)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

// AnnouncementPreview renders the published per-application results of a batch
// as announcement text (结果公示两套实现的桥接).
func (ctrl *ResultController) AnnouncementPreview(c *gin.Context) {
	batchID := parseUint(c.Query("batchId"))
	content, err := ctrl.service.BuildAnnouncementContent(batchID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, gin.H{"content": content})
}

// ListPublishedResults lists the per-application results already made public.
// Managers see every batch; applicants use the /member endpoint below.
func (ctrl *ResultController) ListPublishedResults(c *gin.Context) {
	page, size := getPage(c)
	batchID := parseUint(c.Query("batchId"))
	keyword := c.Query("keyword")
	list, total, err := ctrl.service.ListPublishedResults(page, size, batchID, keyword)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

// ListPublishedAnnouncements lists only published announcements (frontend)
func (ctrl *ResultController) ListPublishedAnnouncements(c *gin.Context) {
	page, size := getPage(c)
	keyword := c.Query("keyword")
	list, total, err := ctrl.service.ListAnnouncements(page, size, keyword, true)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

// --- Certificates (证书) ---

func (ctrl *ResultController) IssueCertificate(c *gin.Context) {
	var cert models.Certificate
	if err := c.ShouldBindJSON(&cert); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.IssueCertificate(&cert); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "证书管理", "颁发证书", "application="+itoa64(cert.ApplicationID))
	response.Ok(c, cert)
}

func (ctrl *ResultController) UpdateCertificate(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.UpdateCertificate(id, updates); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "证书管理", "编辑证书", "id="+c.Param("id"))
	response.OkWithMessage(c, "更新成功", nil)
}

func (ctrl *ResultController) ListCertificates(c *gin.Context) {
	page, size := getPage(c)
	keyword := c.Query("keyword")
	list, total, err := ctrl.service.ListCertificates(page, size, keyword)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

type VoidCertReq struct {
	Reason string `json:"reason"`
}

// VoidCertificate invalidates an issued certificate (作废证书).
func (ctrl *ResultController) VoidCertificate(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req VoidCertReq
	// A missing body is fine: the reason is optional.
	_ = c.ShouldBindJSON(&req)
	if err := ctrl.service.VoidCertificate(id, req.Reason); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "证书管理", "作废证书", "id="+c.Param("id"))
	response.OkWithMessage(c, "证书已作废", nil)
}

// MyCertificates lists the current applicant's certificates (frontend)
func (ctrl *ResultController) MyCertificates(c *gin.Context) {
	userID := middleware.GetUserID(c)
	list, err := ctrl.service.ListUserCertificates(userID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, list)
}

func itoa64(n uint64) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
