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
