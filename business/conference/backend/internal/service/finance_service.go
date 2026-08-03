package service

import (
	"errors"
	"fmt"
	"time"

	"conference/internal/models"
	"conference/pkg/db"

	"github.com/google/uuid"
)

type FinanceService struct{}

// CreateOrder creates a payment order for a meeting
func (s *FinanceService) CreateOrder(meetingID, userID uint64) (*models.Order, error) {
	// Check if already has an unpaid/paid order
	var existing models.Order
	err := db.DB.Where("meeting_id = ? AND user_id = ? AND status IN ?",
		meetingID, userID, []string{models.PayStatusUnpaid, models.PayStatusPaid}).First(&existing).Error
	if err == nil {
		if existing.Status == models.PayStatusPaid {
			return nil, errors.New("已支付")
		}
		return &existing, nil // Return existing unpaid order
	}

	var meeting models.Meeting
	if err := db.DB.First(&meeting, meetingID).Error; err != nil {
		return nil, errors.New("会议不存在")
	}
	if meeting.Fee <= 0 {
		return nil, errors.New("该会议免费，无需支付")
	}

	order := models.Order{
		MeetingID: meetingID,
		UserID:    userID,
		OrderNo:   generateOrderNo(),
		Amount:    meeting.Fee,
		Status:    models.PayStatusUnpaid,
	}
	if err := db.DB.Create(&order).Error; err != nil {
		return nil, errors.New("创建订单失败")
	}
	return &order, nil
}

// PayOrder simulates payment (stub for WeChat/Alipay integration)
func (s *FinanceService) PayOrder(orderID, userID uint64, payMethod string) error {
	var order models.Order
	if err := db.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		return errors.New("订单不存在")
	}
	if order.Status != models.PayStatusUnpaid {
		return errors.New("订单状态异常")
	}

	now := time.Now()
	txnID := fmt.Sprintf("TXN%s", uuid.New().String()[:16])
	return db.DB.Model(&order).Updates(map[string]interface{}{
		"status":         models.PayStatusPaid,
		"pay_method":     payMethod,
		"paid_at":        &now,
		"transaction_id": txnID,
	}).Error
}

// ApplyRefund creates a refund request
func (s *FinanceService) ApplyRefund(orderID, userID uint64, reason string) error {
	var order models.Order
	if err := db.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		return errors.New("订单不存在")
	}
	if order.Status != models.PayStatusPaid {
		return errors.New("订单状态不支持退款")
	}

	// Check existing refund
	var count int64
	db.DB.Model(&models.Refund{}).Where("order_id = ? AND status = ?", orderID, models.RefundStatusPending).Count(&count)
	if count > 0 {
		return errors.New("已有退款申请在处理中")
	}

	refund := models.Refund{
		OrderID: orderID,
		UserID:  userID,
		Amount:  order.Amount,
		Reason:  reason,
		Status:  models.RefundStatusPending,
	}
	if err := db.DB.Create(&refund).Error; err != nil {
		return errors.New("提交退款申请失败")
	}
	db.DB.Model(&order).Update("status", models.PayStatusRefunding)
	return nil
}

// ProcessRefund approves or rejects a refund
func (s *FinanceService) ProcessRefund(refundID, adminID uint64, approved bool, comment string) error {
	var refund models.Refund
	if err := db.DB.First(&refund, refundID).Error; err != nil {
		return errors.New("退款申请不存在")
	}
	if refund.Status != models.RefundStatusPending {
		return errors.New("退款申请已处理")
	}

	now := time.Now()
	status := models.RefundStatusRejected
	if approved {
		status = models.RefundStatusApproved
	}

	if err := db.DB.Model(&refund).Updates(map[string]interface{}{
		"status":         status,
		"reviewed_by":    adminID,
		"reviewed_at":    &now,
		"review_comment": comment,
	}).Error; err != nil {
		return errors.New("处理失败")
	}

	// Update order status
	orderStatus := models.PayStatusPaid
	if approved {
		orderStatus = models.PayStatusRefunded
	}
	db.DB.Model(&models.Order{}).Where("id = ?", refund.OrderID).Update("status", orderStatus)
	return nil
}

// ListOrders returns paginated orders
func (s *FinanceService) ListOrders(meetingID uint64, status string, page, size int) ([]models.Order, int64, error) {
	var list []models.Order
	var total int64
	query := db.DB.Model(&models.Order{}).Preload("User").Preload("Meeting")
	if meetingID > 0 {
		query = query.Where("meeting_id = ?", meetingID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

// ListMyOrders returns user's orders
func (s *FinanceService) ListMyOrders(userID uint64, status string, page, size int) ([]models.Order, int64, error) {
	var list []models.Order
	var total int64
	query := db.DB.Model(&models.Order{}).Preload("Meeting").Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

// ListRefunds returns paginated refund requests
func (s *FinanceService) ListRefunds(status string, page, size int) ([]models.Refund, int64, error) {
	var list []models.Refund
	var total int64
	query := db.DB.Model(&models.Refund{}).Preload("Order").Preload("User")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

// GetLedger returns financial summary per meeting
func (s *FinanceService) GetLedger() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	var meetings []models.Meeting
	db.DB.Find(&meetings)

	for _, m := range meetings {
		var income float64
		var refunded float64
		var orderCount int64

		db.DB.Model(&models.Order{}).Where("meeting_id = ? AND status = ?", m.ID, models.PayStatusPaid).
			Select("COALESCE(SUM(amount), 0)").Scan(&income)
		db.DB.Model(&models.Order{}).Where("meeting_id = ? AND status = ?", m.ID, models.PayStatusRefunded).
			Select("COALESCE(SUM(amount), 0)").Scan(&refunded)
		db.DB.Model(&models.Order{}).Where("meeting_id = ? AND status IN ?", m.ID,
			[]string{models.PayStatusPaid, models.PayStatusRefunded, models.PayStatusRefunding}).Count(&orderCount)

		results = append(results, map[string]interface{}{
			"meeting_id":    m.ID,
			"meeting_title": m.Title,
			"income":        income,
			"refunded":      refunded,
			"net_income":    income - refunded,
			"order_count":   orderCount,
		})
	}
	return results, nil
}

// SaveInvoice saves invoice info
func (s *FinanceService) SaveInvoice(orderID, userID uint64, invoice *models.Invoice) error {
	var order models.Order
	if err := db.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		return errors.New("订单不存在")
	}

	invoice.OrderID = orderID
	invoice.UserID = userID

	// Check existing
	var existing models.Invoice
	err := db.DB.Where("order_id = ?", orderID).First(&existing).Error
	if err == nil {
		return db.DB.Model(&existing).Updates(invoice).Error
	}
	return db.DB.Create(invoice).Error
}

// GetInvoice returns invoice for an order
func (s *FinanceService) GetInvoice(orderID, userID uint64) (*models.Invoice, error) {
	var invoice models.Invoice
	err := db.DB.Where("order_id = ? AND user_id = ?", orderID, userID).First(&invoice).Error
	if err != nil {
		return nil, errors.New("发票信息不存在")
	}
	return &invoice, nil
}

func generateOrderNo() string {
	return fmt.Sprintf("CF%s", time.Now().Format("20060102150405")+uuid.New().String()[:6])
}
