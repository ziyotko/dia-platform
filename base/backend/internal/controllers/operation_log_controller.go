package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"base/internal/service"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

type OperationLogController struct {
	service service.OperationLogService
}

func (ctl *OperationLogController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	if size > 500 {
		size = 500
	}
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))

	q := service.LogListQuery{
		TenantID: c.GetUint64("tenantID"),
		Username: c.Query("username"),
		Module:   c.Query("module"),
		Action:   c.Query("action"),
		Method:   c.Query("method"),
		Path:     c.Query("path"),
		Status:   status,
		StartAt:  c.Query("startAt"),
		EndAt:    c.Query("endAt"),
		Page:     page,
		Size:     size,
	}

	list, total, err := ctl.service.List(q)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctl *OperationLogController) Delete(c *gin.Context) {
	var req struct {
		IDs []uint64 `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	if err := ctl.service.DeleteByIDs(req.IDs); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctl *OperationLogController) Clear(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days <= 0 {
		days = 30
	}
	tenantID := c.GetUint64("tenantID")
	if err := ctl.service.ClearBefore(days, tenantID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "清理成功", nil)
}

func (ctl *OperationLogController) Export(c *gin.Context) {
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	q := service.LogListQuery{
		TenantID: c.GetUint64("tenantID"),
		Username: c.Query("username"),
		Module:   c.Query("module"),
		Action:   c.Query("action"),
		Method:   c.Query("method"),
		Path:     c.Query("path"),
		Status:   status,
		StartAt:  c.Query("startAt"),
		EndAt:    c.Query("endAt"),
	}

	csv, err := ctl.service.Export(q)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	filename := "operation_logs_" + time.Now().Format("20060102_150405") + ".csv"
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.String(http.StatusOK, "\uFEFF"+strings.TrimSuffix(csv, "\n"))
}
