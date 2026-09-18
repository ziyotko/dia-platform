package controllers

import (
	"fmt"

	"member/internal/middleware"
	"member/internal/models"
	"member/internal/service"
	"member/pkg/db"
	"member/pkg/response"

	"github.com/gin-gonic/gin"
)

type CertificateController struct {
	certService service.CertificateService
}

func (ctrl *CertificateController) GetMyCertificates(c *gin.Context) {
	memberID := middleware.GetMemberID(c)
	certs, err := ctrl.certService.GetMyCertificates(memberID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, certs)
}

func (ctrl *CertificateController) GetCertificate(c *gin.Context) {
	memberID := middleware.GetMemberID(c)
	id := parseUint(c.Param("id"))
	cert, err := ctrl.certService.GetCertificate(id, memberID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, cert)
}

func (ctrl *CertificateController) CreateCertificate(c *gin.Context) {
	var req service.CreateCertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	cert, err := ctrl.certService.CreateCertificate(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "创建成功", cert)
}

// ListCertificates lists issued certificates (admin)
func (ctrl *CertificateController) ListCertificates(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 10)
	keyword := c.Query("keyword")
	status := c.Query("status")

	list, total, err := ctrl.certService.ListCertificates(page, size, keyword, status)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// RegenerateCertificate (re)generates the certificate PDF file (admin)
func (ctrl *CertificateController) RegenerateCertificate(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var cert models.Certificate
	if err := db.DB.First(&cert, id).Error; err != nil {
		response.NotFound(c, "证书不存在")
		return
	}
	if err := ctrl.certService.GenerateFileForCertificate(&cert); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "证书文件已生成", cert)
}

// RegenerateMissingCertificates 批量补生成历史存量中缺失的证书 PDF（admin）
func (ctrl *CertificateController) RegenerateMissingCertificates(c *gin.Context) {
	limit := parseIntDefault(c.Query("limit"), 200)
	ok, failed, failures, err := ctrl.certService.RegenerateMissingCertificates(limit)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, fmt.Sprintf("已生成 %d 张，失败 %d 张", ok, failed),
		gin.H{"ok": ok, "failed": failed, "failures": failures})
}

func (ctrl *CertificateController) UpdateCertificate(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req struct {
		FilePath *string `json:"file_path"`
		Status   *string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.certService.UpdateCertificate(id, req.FilePath, req.Status); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "更新成功", nil)
}

// RenewCertificate renews the current member's certificate (expires old ones, creates new)
func (ctrl *CertificateController) RenewCertificate(c *gin.Context) {
	memberID := middleware.GetMemberID(c)
	cert, err := ctrl.certService.RenewMyCertificate(memberID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "证书已重新生成", cert)
}
