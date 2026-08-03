package controllers

import (
	"strconv"

	"conference/internal/middleware"
	"conference/internal/models"
	"conference/internal/service"
	"conference/pkg/response"

	"github.com/gin-gonic/gin"
)

type FinanceController struct {
	service service.FinanceService
}

// --- Member endpoints ---

func (ctrl *FinanceController) CreateOrder(c *gin.Context) {
	meetingID := parseUint(c.Param("meetingId"))
	userID := middleware.GetUserID(c)

	order, err := ctrl.service.CreateOrder(meetingID, userID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, order)
}

func (ctrl *FinanceController) PayOrder(c *gin.Context) {
	orderID := parseUint(c.Param("id"))
	userID := middleware.GetUserID(c)
	var req struct {
		PayMethod string `json:"payMethod" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请选择支付方式")
		return
	}
	if err := ctrl.service.PayOrder(orderID, userID, req.PayMethod); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "支付成功", nil)
}

func (ctrl *FinanceController) ApplyRefund(c *gin.Context) {
	orderID := parseUint(c.Param("id"))
	userID := middleware.GetUserID(c)
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写退款原因")
		return
	}
	if err := ctrl.service.ApplyRefund(orderID, userID, req.Reason); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "退款申请已提交", nil)
}

func (ctrl *FinanceController) MyOrders(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	status := c.Query("status")

	list, total, err := ctrl.service.ListMyOrders(userID, status, page, size)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctrl *FinanceController) SaveInvoice(c *gin.Context) {
	orderID := parseUint(c.Param("id"))
	userID := middleware.GetUserID(c)
	var invoice models.Invoice
	if err := c.ShouldBindJSON(&invoice); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.SaveInvoice(orderID, userID, &invoice); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "保存成功", nil)
}

func (ctrl *FinanceController) GetInvoice(c *gin.Context) {
	orderID := parseUint(c.Param("id"))
	userID := middleware.GetUserID(c)

	invoice, err := ctrl.service.GetInvoice(orderID, userID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, invoice)
}

// --- Admin endpoints ---

func (ctrl *FinanceController) ListOrders(c *gin.Context) {
	meetingID := parseUint(c.Query("meetingId"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	status := c.Query("status")

	list, total, err := ctrl.service.ListOrders(meetingID, status, page, size)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctrl *FinanceController) ListRefunds(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	status := c.Query("status")

	list, total, err := ctrl.service.ListRefunds(status, page, size)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctrl *FinanceController) ProcessRefund(c *gin.Context) {
	id := parseUint(c.Param("id"))
	adminID := middleware.GetAdminID(c)
	var req struct {
		Approved bool   `json:"approved"`
		Comment  string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.ProcessRefund(id, adminID, req.Approved, req.Comment); err != nil {
		response.Fail(c, err.Error())
		return
	}
	msg := "退款已拒绝"
	if req.Approved {
		msg = "退款已通过"
	}
	response.OkWithMessage(c, msg, nil)
}

func (ctrl *FinanceController) GetLedger(c *gin.Context) {
	ledger, err := ctrl.service.GetLedger()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, ledger)
}
