package controllers

import (
	"errors"
	"member/config"
	"member/internal/middleware"
	"member/internal/service"
	"member/pkg/response"
	"member/pkg/utils"
	"net/http"
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
		response.ServerErrorFrom(c, err)
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
		response.BadRequestFrom(c, err)
		return
	}
	response.SuccessWithMessage(c, "缴费信息已提交，等待管理员确认", nil)
}

// ConfirmFee confirms a pending fee (admin)
func (ctrl *FeeController) ConfirmFee(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req struct {
		Amount float64 `json:"amount"`
		Remark *string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.feeService.ConfirmFee(id, req.Amount, req.Remark, middleware.GetUsername(c)); err != nil {
		response.BadRequestFrom(c, err)
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
		response.BadRequestFrom(c, err)
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
		response.BadRequestFrom(c, err)
		return
	}
	response.SuccessWithMessage(c, "开票申请已提交", nil)
}

// GetMemberFeeInfo returns member's org and level info for fee creation
func (ctrl *FeeController) GetMemberFeeInfo(c *gin.Context) {
	memberID := parseUint(c.Param("id"))
	orgID, levelID, orgName, levelName, err := ctrl.feeService.GetMemberFeeInfo(memberID)
	if err != nil {
		response.BadRequestFrom(c, err)
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
	var req service.UpdateFeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.feeService.UpdateFeeRecord(id, req, middleware.GetUsername(c)); err != nil {
		response.BadRequestFrom(c, err)
		return
	}
	response.SuccessWithMessage(c, "更新成功", nil)
}

func (ctrl *FeeController) DeleteFee(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.feeService.DeleteFeeRecord(id); err != nil {
		response.BadRequestFrom(c, err)
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}

// IssueInvoice issues an invoice (admin uploads PDF, sets invoice_no, marks as issued)
func (ctrl *FeeController) IssueInvoice(c *gin.Context) {
	id := parseUint(c.Param("id"))

	// 解析前先限制请求体总大小（同 /upload）
	limitMultipartBody(c, maxUploadSize)

	invoiceNo := c.PostForm("invoice_no")
	if invoiceNo == "" {
		response.BadRequest(c, "请输入票据号码")
		return
	}

	var invoiceFile string
	file, err := c.FormFile("file")
	switch {
	case err == nil:
		if file.Size > maxUploadSize {
			response.BadRequest(c, "文件大小不能超过 10MB")
			return
		}
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
		// 与 /upload 保持一致：返回带部署前缀的绝对路径，
		// 否则前端（及静态挂载 <upload_dir_prefix>/uploads）会 404
		invoiceFile = config.Cfg.Server.UploadDirPrefix + "/" + path
	case errors.Is(err, http.ErrMissingFile):
		// 未附发票文件，允许只更新票据号
	default:
		// 超限或解析失败：不能当成「没传文件」静默继续（原先会跳过文件直接开票）
		if isBodyTooLarge(err) {
			response.BadRequest(c, "文件大小不能超过 10MB")
			return
		}
		response.BadRequest(c, "文件上传失败")
		return
	}

	if err := ctrl.feeService.IssueInvoice(id, invoiceNo, invoiceFile); err != nil {
		response.BadRequestFrom(c, err)
		return
	}
	response.SuccessWithMessage(c, "已开票", nil)
}

func (ctrl *FeeController) ListAllFees(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parsePageSize(c)
	year := parseIntDefault(c.Query("year"), 0)
	status := c.Query("status")
	memberType := c.Query("member_type")
	memberIDStr := c.Query("member_id")
	var memberID uint64
	if memberIDStr != "" {
		memberID = parseUint(memberIDStr)
	}

	fees, total, err := ctrl.feeService.ListAllFees(page, size, year, status, memberID, memberType)
	if err != nil {
		response.ServerErrorFrom(c, err)
		return
	}
	// 统计卡片数据：按年份/会员类型过滤的状态分布与金额合计。
	// 刻意不叠加 status 筛选（卡片本身就是状态拆分，叠加后未缴费/待确认会恒为 0），
	// 也不受分页影响（原实现由前端按"当前页"计算，翻页数字会跳变）。
	summary, err := ctrl.feeService.GetFeeListSummary(year, memberType)
	if err != nil {
		response.ServerErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"list":    fees,
		"total":   total,
		"page":    page,
		"size":    size,
		"summary": summary,
	})
}
