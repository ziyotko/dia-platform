package service

import (
	"errors"
	"member/internal/models"
	"member/pkg/db"
	"time"
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

// PayFee submits payment info (receipt + date) and sets status to pending
func (s *FeeService) PayFee(memberID, feeID uint64, receiptFile, paidDate string) error {
	var fee models.FeeRecord
	if err := db.DB.Where("id = ? AND member_id = ?", feeID, memberID).First(&fee).Error; err != nil {
		return errors.New("费用记录不存在")
	}
	if fee.Status != models.FeeStatusUnpaid {
		return errors.New("该费用已提交，请等待管理员确认")
	}

	updates := map[string]interface{}{
		"status":       models.FeeStatusPending,
		"receipt_file": receiptFile,
		"paid_date":    paidDate,
	}
	if err := db.DB.Model(&fee).Updates(updates).Error; err != nil {
		return err
	}
	return nil
}

// ConfirmFee confirms a pending fee (admin)
func (s *FeeService) ConfirmFee(id uint64, amount float64, remark string) error {
	var fee models.FeeRecord
	if err := db.DB.First(&fee, id).Error; err != nil {
		return errors.New("费用记录不存在")
	}
	if fee.Status != models.FeeStatusPending {
		return errors.New("只有待确认的费用才能确认缴费")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       models.FeeStatusPaid,
		"paid_at":      &models.LocalTime{Time: now},
		"confirmed_at": &models.LocalTime{Time: now},
		"paid_amount":  amount,
	}
	if amount > 0 {
		updates["paid_amount"] = amount
	}
	if remark != "" {
		updates["remark"] = remark
	}
	if err := db.DB.Model(&fee).Updates(updates).Error; err != nil {
		return err
	}

	// Update member status to active if pending_payment
	db.DB.Model(&models.Member{}).Where("id = ? AND status = ?", fee.MemberID, models.MemberStatusPendingPay).
		Update("status", models.MemberStatusActive)

	return nil
}

// CreateFeeRecord creates a fee record (admin)
func (s *FeeService) CreateFeeRecord(req CreateFeeRequest) (*models.FeeRecord, error) {
	// Check duplicate: same member + same year
	var count int64
	db.DB.Model(&models.FeeRecord{}).Where("member_id = ? AND year = ?", req.MemberID, req.Year).Count(&count)
	if count > 0 {
		return nil, errors.New("该会员本年度费用记录已存在")
	}

	fee := models.FeeRecord{
		MemberID:  req.MemberID,
		Year:      req.Year,
		Amount:    req.Amount,
		Status:    models.FeeStatusUnpaid,
		Remark:    req.Remark,
		OrgID:     req.OrgID,
		OrgName:   req.OrgName,
		LevelID:   req.LevelID,
		LevelName: req.LevelName,
	}
	if err := db.DB.Create(&fee).Error; err != nil {
		return nil, err
	}
	return &fee, nil
}

// UpdateFeeRecord updates a fee record (admin)
func (s *FeeService) UpdateFeeRecord(id uint64, status, invoiceNo string, amount float64, remark string, levelID uint64, levelName string) error {
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
	if levelID > 0 {
		updates["level_id"] = levelID
	}
	if levelName != "" {
		updates["level_name"] = levelName
	}

	if err := db.DB.Model(&models.FeeRecord{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return err
	}

	// If confirmed as paid, activate member
	if status == models.FeeStatusPaid {
		var fee models.FeeRecord
		db.DB.First(&fee, id)
		db.DB.Model(&models.Member{}).Where("id = ? AND status = ?", fee.MemberID, models.MemberStatusPendingPay).
			Update("status", models.MemberStatusActive)
	}

	return nil
}

// DeleteFeeRecord deletes a fee record (admin, unpaid only)
func (s *FeeService) DeleteFeeRecord(id uint64) error {
	var fee models.FeeRecord
	if err := db.DB.First(&fee, id).Error; err != nil {
		return errors.New("费用记录不存在")
	}
	if fee.Status == models.FeeStatusPaid {
		return errors.New("已缴费记录不能删除")
	}
	return db.DB.Delete(&fee).Error
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
	MemberID  uint64  `json:"member_id" binding:"required"`
	Year      int     `json:"year" binding:"required"`
	Amount    float64 `json:"amount" binding:"required"`
	Remark    string  `json:"remark"`
	OrgID     uint64  `json:"org_id"`
	OrgName   string  `json:"org_name"`
	LevelID   uint64  `json:"level_id"`
	LevelName string  `json:"level_name"`
}

// ApplyInvoice submits an invoice application for a paid fee record (member)
func (s *FeeService) ApplyInvoice(memberID, feeID uint64, req ApplyInvoiceRequest) error {
	var fee models.FeeRecord
	if err := db.DB.Where("id = ? AND member_id = ?", feeID, memberID).First(&fee).Error; err != nil {
		return errors.New("费用记录不存在")
	}
	if fee.Status != models.FeeStatusPaid {
		return errors.New("只有已缴费的费用才能申请开票")
	}
	if fee.InvoiceStatus != "" {
		return errors.New("该费用已申请开票，请勿重复申请")
	}

	updates := map[string]interface{}{
		"invoice_status":  "applied",
		"invoice_company": req.InvoiceCompany,
		"invoice_tax_id":  req.InvoiceTaxID,
		"invoice_amount":  req.InvoiceAmount,
		"invoice_contact": req.InvoiceContact,
		"invoice_remark":  req.InvoiceRemark,
	}
	return db.DB.Model(&fee).Updates(updates).Error
}

// IssueInvoice issues an invoice for an applied fee record (admin)
// Also supports re-uploading invoice file when already issued
func (s *FeeService) IssueInvoice(id uint64, invoiceNo, invoiceFile string) error {
	var fee models.FeeRecord
	if err := db.DB.First(&fee, id).Error; err != nil {
		return errors.New("费用记录不存在")
	}
	if fee.InvoiceStatus != "applied" && fee.InvoiceStatus != "issued" {
		return errors.New("只有已申请开票或已开票的费用才能操作")
	}

	updates := map[string]interface{}{}
	if fee.InvoiceStatus == "applied" {
		updates["invoice_status"] = "issued"
		updates["invoice_issued_at"] = &models.LocalTime{Time: time.Now()}
	}
	updates["invoice_no"] = invoiceNo
	if invoiceFile != "" {
		updates["invoice_file"] = invoiceFile
	}
	return db.DB.Model(&fee).Updates(updates).Error
}

// GetMemberFeeInfo returns a member's org and level info from their approved application
func (s *FeeService) GetMemberFeeInfo(memberID uint64) (orgID, levelID uint64, orgName, levelName string, err error) {
	var app models.Application
	if err := db.DB.Preload("Org").Where("member_id = ? AND status = ?", memberID, models.AppStatusApproved).
		Order("created_at DESC").First(&app).Error; err != nil {
		return 0, 0, "", "", errors.New("未找到该会员的入会申请")
	}

	orgID = app.OrgID
	if app.Org.ID > 0 {
		orgName = app.Org.Name
	}

	// Look up the member's level from their org's lowest level
	var orgLevel models.MemberOrgLevel
	if err := db.DB.Preload("Level").
		Joins("JOIN member_levels ml ON ml.id = member_org_levels.level_id").
		Where("member_org_levels.org_id = ?", app.OrgID).
		Order("ml.level ASC").
		First(&orgLevel).Error; err == nil {
		levelID = orgLevel.LevelID
		levelName = orgLevel.Level.Name
	}

	return orgID, levelID, orgName, levelName, nil
}

// ApplyInvoiceRequest represents invoice application data submitted by member
type ApplyInvoiceRequest struct {
	InvoiceCompany string  `json:"invoice_company" binding:"required"`
	InvoiceTaxID   string  `json:"invoice_tax_id" binding:"required"`
	InvoiceAmount  float64 `json:"invoice_amount" binding:"required"`
	InvoiceContact string  `json:"invoice_contact" binding:"required"`
	InvoiceRemark  string  `json:"invoice_remark"`
}
