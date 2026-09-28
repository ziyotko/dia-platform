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

// MaxPageSize 单页最大条数。
// 注意：超出上限时收敛到上限，**不要**退回默认值——早期实现是 `size > 100 → 10`，
// 于是前端传 pageSize=200 的调用（批次/申报人下拉框）会被悄悄截断成 10 条。
const MaxPageSize = 100

func getPage(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	if size > MaxPageSize {
		size = MaxPageSize
	}
	return page, size
}

func recordAudit(c *gin.Context, module, action, detail string) {
	svc := service.AuditService{}
	svc.Record(middleware.GetAdminID(c), middleware.GetAdminUsername(c), module, action, c.ClientIP(), detail)
}
