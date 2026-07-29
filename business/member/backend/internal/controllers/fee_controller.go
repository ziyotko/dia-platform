package controllers

import (
	"member/internal/middleware"
	"member/internal/service"
	"member/pkg/response"
	"member/pkg/utils"
	"strings"

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

	var req struct {
		ReceiptFile string `json:"receipt_file"`
		PaidDate    string `json:"paid_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := ctrl.feeService.PayFee(memberID, feeID, req.ReceiptFile, req.PaidDate); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "缴费信息已提交，等待管理员确认", nil)
}

// ConfirmFee confirms a pending fee (admin)
func (ctrl *FeeController) ConfirmFee(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req struct {
		Amount float64 `json:"amount"`
		Remark string  `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.feeService.ConfirmFee(id, req.Amount, req.Remark); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "已确认缴费", nil)
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

func (ctrl *FeeController) ApplyInvoice(c *gin.Context) {
	memberID := middleware.GetMemberID(c)
	feeID := parseUint(c.Param("id"))

	var req service.ApplyInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整开票信息")
		return
	}

	if err := ctrl.feeService.ApplyInvoice(memberID, feeID, req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "开票申请已提交", nil)
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
		LevelID   uint64  `json:"level_id"`
		LevelName string  `json:"level_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.feeService.UpdateFeeRecord(id, req.Status, req.InvoiceNo, req.Amount, req.Remark, req.LevelID, req.LevelName); err != nil {
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

// IssueInvoice issues an invoice (admin uploads PDF, sets invoice_no, marks as issued)
func (ctrl *FeeController) IssueInvoice(c *gin.Context) {
	id := parseUint(c.Param("id"))

	invoiceNo := c.PostForm("invoice_no")
	if invoiceNo == "" {
		response.BadRequest(c, "请输入票据号码")
		return
	}

	var invoiceFile string
	file, err := c.FormFile("file")
	if err == nil {
		// Check file type
		ext := ""
		if idx := strings.LastIndex(file.Filename, "."); idx >= 0 {
			ext = strings.ToLower(file.Filename[idx:])
		}
		if ext != ".pdf" {
			response.BadRequest(c, "仅支持 PDF 格式")
			return
		}
		path, err := utils.SaveUploadedFile(file, "invoices")
		if err != nil {
			response.ServerError(c, "文件上传失败")
			return
		}
		invoiceFile = "/" + path
	}

	if err := ctrl.feeService.IssueInvoice(id, invoiceNo, invoiceFile); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "已开票", nil)
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
