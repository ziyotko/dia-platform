package controllers

import (
	"member/internal/service"
	"member/pkg/response"

	"github.com/gin-gonic/gin"
)

type OperationLogController struct {
	logService service.OperationLogService
}

// List lists operation logs (admin)
func (ctrl *OperationLogController) List(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 10)
	keyword := c.Query("keyword")

	list, total, err := ctrl.logService.List(page, size, keyword)
	if err != nil {
		response.ServerErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  size,
	})
}
