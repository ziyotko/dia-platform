package controllers

import (
	"strconv"

	"application/internal/middleware"
	"application/internal/service"

	"github.com/gin-gonic/gin"
)

func parseUint(s string) uint64 {
	v, _ := strconv.ParseUint(s, 10, 64)
	return v
}

func getPage(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	return page, size
}

func recordAudit(c *gin.Context, module, action, detail string) {
	svc := service.AuditService{}
	svc.Record(middleware.GetAdminID(c), middleware.GetAdminUsername(c), module, action, c.ClientIP(), detail)
}
