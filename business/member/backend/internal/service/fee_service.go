package service

import (
	"errors"
	"member/internal/models"
	"member/pkg/db"
	"time"

	"github.com/google/uuid"
)

type FeeService struct{}

// GetMyFees returns the member's fee records
func (s *FeeService) GetMyFees(memberID uint64, year int, status string) ([]models.FeeRecord, error) {
	var fees []models.FeeRecord
	query := db.DB.Where("member_id = ?", memberID)
	if year > 0 {
		query = query.Where("year = ?", year)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Order("year DESC").Find(&fees).Error; err != nil {
		return nil, err
	}
	return fees, nil
}

// PayFee simulates fee payment
func (s *FeeService) PayFee(memberID, feeID uint64) error {
	var fee models.FeeRecord
	if err := db.DB.Where("id = ? AND member_id = ?", feeID, memberID).First(&fee).Error; err != nil {
		return errors.New("费用记录不存在")
	}
	if fee.Status == models.FeeStatusPaid {
		return errors.New("该费用已缴纳")
	}

	now := time.Now()
	txID := "TXN-" + uuid.New().String()[:8]

	updates := map[string]interface{}{
		"status":         models.FeeStatusPaid,
		"paid_at":        &models.LocalTime{Time: now},
		"transaction_id": txID,
	}
	if err := db.DB.Model(&fee).Updates(updates).Error; err != nil {
		return err
	}

	// Update member status to active if pending_payment
	db.DB.Model(&models.Member{}).Where("id = ? AND status = ?", memberID, models.MemberStatusPendingPay).
		Update("status", models.MemberStatusActive)

	return nil
}

// CreateFeeRecord creates a fee record (admin)
func (s *FeeService) CreateFeeRecord(req CreateFeeRequest) (*models.FeeRecord, error) {
	fee := models.FeeRecord{
		MemberID: req.MemberID,
		Year:     req.Year,
		Amount:   req.Amount,
		Status:   models.FeeStatusUnpaid,
		Remark:   req.Remark,
	}
	if err := db.DB.Create(&fee).Error; err != nil {
		return nil, err
	}
	return &fee, nil
}

// UpdateFeeRecord updates a fee record (admin)
func (s *FeeService) UpdateFeeRecord(id uint64, status, invoiceNo string, amount float64, remark string) error {
	updates := map[string]interface{}{}
	if status != "" {
		updates["status"] = status
		if status == models.FeeStatusPaid {
			now := time.Now()
			updates["paid_at"] = &models.LocalTime{Time: now}
		}
	}
	if invoiceNo != "" {
		updates["invoice_no"] = invoiceNo
	}
	if amount > 0 {
		updates["amount"] = amount
	}
	if remark != "" {
		updates["remark"] = remark
	}
	return db.DB.Model(&models.FeeRecord{}).Where("id = ?", id).Updates(updates).Error
}

// ListAllFees lists all fee records (admin)
func (s *FeeService) ListAllFees(page, size int, year int, status string, memberID uint64) ([]models.FeeRecord, int64, error) {
	var fees []models.FeeRecord
	var total int64

	query := db.DB.Model(&models.FeeRecord{}).Preload("Member")
	if year > 0 {
		query = query.Where("year = ?", year)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if memberID > 0 {
		query = query.Where("member_id = ?", memberID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("year DESC").Offset((page - 1) * size).Limit(size).Find(&fees).Error; err != nil {
		return nil, 0, err
	}
	return fees, total, nil
}

type CreateFeeRequest struct {
	MemberID uint64  `json:"member_id" binding:"required"`
	Year     int     `json:"year" binding:"required"`
	Amount   float64 `json:"amount" binding:"required"`
	Remark   string  `json:"remark"`
}
