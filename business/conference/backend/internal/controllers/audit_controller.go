package controllers

import (
	"strconv"

	"conference/internal/service"
	"conference/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuditController struct {
	service service.AuditService
}

func (ctrl *AuditController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	userType := c.Query("userType")
	action := c.Query("action")
	resource := c.Query("resource")

	list, total, err := ctrl.service.List(userType, action, resource, page, size)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}
