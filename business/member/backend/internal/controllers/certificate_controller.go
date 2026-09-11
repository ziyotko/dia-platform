package controllers

import (
	"member/internal/middleware"
	"member/internal/service"
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
