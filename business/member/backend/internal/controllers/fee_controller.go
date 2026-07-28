package controllers

import (
	"member/internal/middleware"
	"member/internal/service"
	"member/pkg/response"

	"github.com/gin-gonic/gin"
)

type FeeController struct {
	feeService service.FeeService
}

func (ctrl *FeeController) GetMyFees(c *gin.Context) {
	memberID := middleware.GetMemberID(c)
	year := parseIntDefault(c.Query("year"), 0)
	status := c.Query("status")

	fees, err := ctrl.feeService.GetMyFees(memberID, year, status)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, fees)
}

func (ctrl *FeeController) PayFee(c *gin.Context) {
	memberID := middleware.GetMemberID(c)
	feeID := parseUint(c.Param("id"))

	if err := ctrl.feeService.PayFee(memberID, feeID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "缴费成功", nil)
}

func (ctrl *FeeController) CreateFee(c *gin.Context) {
	var req service.CreateFeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	fee, err := ctrl.feeService.CreateFeeRecord(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "创建成功", fee)
}

// GetMemberFeeInfo returns member's org and level info for fee creation
func (ctrl *FeeController) GetMemberFeeInfo(c *gin.Context) {
	memberID := parseUint(c.Param("id"))
	orgID, levelID, orgName, levelName, err := ctrl.feeService.GetMemberFeeInfo(memberID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"org_id":     orgID,
		"level_id":   levelID,
		"org_name":   orgName,
		"level_name": levelName,
	})
}

func (ctrl *FeeController) UpdateFee(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req struct {
		Status    string  `json:"status"`
		InvoiceNo string  `json:"invoice_no"`
		Amount    float64 `json:"amount"`
		Remark    string  `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.feeService.UpdateFeeRecord(id, req.Status, req.InvoiceNo, req.Amount, req.Remark); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "更新成功", nil)
}

func (ctrl *FeeController) DeleteFee(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.feeService.DeleteFeeRecord(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}

func (ctrl *FeeController) ListAllFees(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 10)
	year := parseIntDefault(c.Query("year"), 0)
	status := c.Query("status")
	memberIDStr := c.Query("member_id")
	var memberID uint64
	if memberIDStr != "" {
		memberID = parseUint(memberIDStr)
	}

	fees, total, err := ctrl.feeService.ListAllFees(page, size, year, status, memberID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"list":  fees,
		"total": total,
		"page":  page,
		"size":  size,
	})
}
