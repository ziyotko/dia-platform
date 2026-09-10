package controllers

import (
	"application/internal/service"
	"application/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuditController struct {
	service service.AuditService
}

func (ctrl *AuditController) List(c *gin.Context) {
	page, size := getPage(c)
	module := c.Query("module")
	keyword := c.Query("keyword")
	list, total, err := ctrl.service.List(page, size, module, keyword)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}
