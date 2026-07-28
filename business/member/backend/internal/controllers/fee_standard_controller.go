package controllers

import (
	"member/internal/service"
	"member/pkg/response"

	"github.com/gin-gonic/gin"
)

type FeeStandardController struct {
	feeStdService service.FeeStandardService
}

// ListAll returns all fee standards
func (ctrl *FeeStandardController) ListAll(c *gin.Context) {
	list, err := ctrl.feeStdService.ListAll()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, list)
}

// ListByLevel returns fee standards for a specific level
func (ctrl *FeeStandardController) ListByLevel(c *gin.Context) {
	levelID := parseUint(c.Param("levelId"))
	list, err := ctrl.feeStdService.ListByLevel(levelID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, list)
}

// Upsert creates or updates a single fee standard
func (ctrl *FeeStandardController) Upsert(c *gin.Context) {
	var req struct {
		LevelID uint64  `json:"level_id" binding:"required"`
		Year    int     `json:"year" binding:"required"`
		Amount  float64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	fs, err := ctrl.feeStdService.Upsert(req.LevelID, req.Year, req.Amount)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "保存成功", fs)
}

// BatchUpsert creates or updates multiple fee standards for a level at once
func (ctrl *FeeStandardController) BatchUpsert(c *gin.Context) {
	var req service.FeeStandardUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.feeStdService.BatchUpsert(req.LevelID, req.Items); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "批量保存成功", nil)
}

// Delete removes a fee standard
func (ctrl *FeeStandardController) Delete(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.feeStdService.Delete(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}
