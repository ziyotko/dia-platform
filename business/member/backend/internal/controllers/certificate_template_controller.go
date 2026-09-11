package controllers

import (
	"member/internal/service"
	"member/pkg/response"

	"github.com/gin-gonic/gin"
)

type CertificateTemplateController struct {
	tplService service.CertificateTemplateService
}

// List returns all certificate templates
func (ctrl *CertificateTemplateController) List(c *gin.Context) {
	list, err := ctrl.tplService.List()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, list)
}

// Get returns a single template
func (ctrl *CertificateTemplateController) Get(c *gin.Context) {
	id := parseUint(c.Param("id"))
	tpl, err := ctrl.tplService.Get(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, tpl)
}

// Create creates a new certificate template
func (ctrl *CertificateTemplateController) Create(c *gin.Context) {
	var req service.CertTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}
	tpl, err := ctrl.tplService.Create(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "创建成功", tpl)
}

// Update updates a certificate template
func (ctrl *CertificateTemplateController) Update(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req service.UpdateCertTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误："+err.Error())
		return
	}
	if err := ctrl.tplService.Update(id, req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "更新成功", nil)
}

// Delete deletes a certificate template
func (ctrl *CertificateTemplateController) Delete(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.tplService.Delete(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}
