package service

import (
	"errors"
	"member/internal/models"
	"member/pkg/db"
	"strconv"
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
	if fee.LevelID == 0 && fee.LevelName == "" {
		return errors.New("会员级别为空，请先联系管理员修改会员级别")
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
func (s *FeeService) ConfirmFee(id uint64, amount float64, remark *string, operator string) error {
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
	if remark != nil {
		updates["remark"] = *remark
	}
	if err := db.DB.Model(&fee).Updates(updates).Error; err != nil {
		return err
	}

	// Update member status to active if pending_payment
	db.DB.Model(&models.Member{}).Where("id = ? AND status = ?", fee.MemberID, models.MemberStatusPendingPay).
		Update("status", models.MemberStatusActive)

	// 确认缴费：会员加入分会/代表处时，若尚未加入总会（根组织），则自动补一条加入总会的记录
	s.ensureRootOrgMembership(fee.MemberID, fee.OrgID, fee.LevelID)

	// Update the member's level on the user record
	if fee.LevelID > 0 {
		db.DB.Model(&models.Member{}).Where("id = ?", fee.MemberID).
			Update("member_level", strconv.FormatUint(fee.LevelID, 10))
	}

	// Update the member's active certificate with level info and template ID
	s.updateCertificateWithLevelAndTemplate(fee.MemberID, fee.LevelID, fee.LevelName)

	// Record membership change (缴费确认)
	s.recordPaymentChange(fee, operator)

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
func (s *FeeService) UpdateFeeRecord(id uint64, req UpdateFeeRequest, operator string) error {
	var exist models.FeeRecord
	if err := db.DB.First(&exist, id).Error; err != nil {
		return errors.New("费用记录不存在")
	}

	updates := map[string]interface{}{}
	if req.Status != nil {
		updates["status"] = *req.Status
		if *req.Status == models.FeeStatusPaid {
			now := time.Now()
			updates["paid_at"] = &models.LocalTime{Time: now}
		}
	}
	if req.InvoiceNo != nil {
		updates["invoice_no"] = *req.InvoiceNo
	}
	if req.Amount != nil {
		updates["amount"] = *req.Amount
	}
	if req.Remark != nil {
		updates["remark"] = *req.Remark
	}
	if req.LevelID != nil {
		updates["level_id"] = *req.LevelID
	}
	if req.LevelName != nil {
		updates["level_name"] = *req.LevelName
	}
	if len(updates) == 0 {
		return nil
	}

	if err := db.DB.Model(&models.FeeRecord{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return err
	}

	// If confirmed as paid, activate member and update certificate
	if req.Status != nil && *req.Status == models.FeeStatusPaid {
		var fee models.FeeRecord
		if err := db.DB.First(&fee, id).Error; err != nil {
			return err
		}
		db.DB.Model(&models.Member{}).Where("id = ? AND status = ?", fee.MemberID, models.MemberStatusPendingPay).
			Update("status", models.MemberStatusActive)

		// 免缴确认：会员加入分会/代表处时，若尚未加入总会（根组织），则自动补一条加入总会的记录
		s.ensureRootOrgMembership(fee.MemberID, fee.OrgID, fee.LevelID)

		// Update the member's level on the user record
		if fee.LevelID > 0 {
			db.DB.Model(&models.Member{}).Where("id = ?", fee.MemberID).
				Update("member_level", strconv.FormatUint(fee.LevelID, 10))
		}

		s.updateCertificateWithLevelAndTemplate(fee.MemberID, fee.LevelID, fee.LevelName)

		// Record membership change (缴费确认)
		s.recordPaymentChange(fee, operator)
	}

	return nil
}

// recordPaymentChange inserts a membership change record for a paid fee confirmation.
// 变更原因为“缴费确认”，原始会籍 id/name 为空，新会籍取自费用记录。
func (s *FeeService) recordPaymentChange(fee models.FeeRecord, operator string) {
	var member models.Member
	if err := db.DB.First(&member, fee.MemberID).Error; err != nil {
		return
	}

	change := models.MemberLevelChange{
		MemberID:     member.ID,
		Username:     member.Username,
		MemberName:   memberDisplayName(&member),
		MemberType:   member.MemberType,
		ChangeYear:   time.Now().Year(),
		OrgID:        fee.OrgID,
		OrgName:      fee.OrgName,
		OldLevelID:   0,
		OldLevelName: "",
		NewLevelID:   fee.LevelID,
		NewLevelName: fee.LevelName,
		Reason:       "缴费确认",
		Operator:     operator,
	}
	db.DB.Create(&change)
}

// ensureRootOrgMembership 在免缴确认时调用：当会员加入的是分会/代表处（非总会）时，
// 检查其是否已加入总会（组织机构中的根组织，parent_id=0）。若未加入，则自动补一条加入总会的记录。
func (s *FeeService) ensureRootOrgMembership(memberID, feeOrgID, levelID uint64) {
	var roots []models.Organization
	if err := db.DB.Where("parent_id = ?", 0).Find(&roots).Error; err != nil || len(roots) == 0 {
		return
	}
	for _, root := range roots {
		if root.ID == 0 || feeOrgID == root.ID {
			continue // 费用本身即针对总会，或未找到根组织
		}
		// 已通过直接加入记录加入总会
		var count int64
		db.DB.Model(&models.MemberOrganization{}).
			Where("member_id = ? AND org_id = ?", memberID, root.ID).Count(&count)
		if count > 0 {
			continue
		}
		// 已通过入会申请（已通过）加入总会
		db.DB.Model(&models.Application{}).
			Where("member_id = ? AND org_id = ? AND status = ?", memberID, root.ID, models.AppStatusApproved).Count(&count)
		if count > 0 {
			continue
		}
		// 手动加入总会
		db.DB.Create(&models.MemberOrganization{
			MemberID: memberID,
			OrgID:    root.ID,
			LevelID:  levelID,
			JoinedAt: time.Now(),
		})
	}
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
func (s *FeeService) ListAllFees(page, size int, year int, status string, memberID uint64, memberType string) ([]models.FeeRecord, int64, error) {
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
	if memberType != "" {
		query = query.Where("member_id IN (SELECT id FROM member_users WHERE member_type = ?)", memberType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&fees).Error; err != nil {
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

// UpdateFeeRequest 会费记录更新请求。
// 字段使用指针，以区分“未提供”（nil）与“清空”（指向零值），
// 否则编辑时无法把备注、票据号等清空。
type UpdateFeeRequest struct {
	Status    *string  `json:"status"`
	InvoiceNo *string  `json:"invoice_no"`
	Amount    *float64 `json:"amount"`
	Remark    *string  `json:"remark"`
	LevelID   *uint64  `json:"level_id"`
	LevelName *string  `json:"level_name"`
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
	if fee.PaidAmount <= 0 {
		return errors.New("实缴金额为0，无法申请开票")
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

// updateCertificateWithLevelAndTemplate updates the member's active certificate with level info and template ID.
// 仅当费用关联了有效等级时才更新证书等级，避免把证书已有的 level_id / level_name / cert_template_id 清空。
func (s *FeeService) updateCertificateWithLevelAndTemplate(memberID, levelID uint64, levelName string) {
	// Find the active certificate for this member
	var cert models.Certificate
	if err := db.DB.Where("member_id = ? AND status = ?", memberID, models.CertStatusActive).First(&cert).Error; err != nil {
		return // no active certificate found, skip
	}

	// 无有效等级（如手工新增的免缴费用）时保持证书原等级不变
	if levelID == 0 {
		return
	}

	updates := map[string]interface{}{
		"level_id":   levelID,
		"level_name": levelName,
	}

	// Look up certificate template for this level; if not found, fall back to the lowest level's template
	var tpl models.MemberCertificateTemplate
	if err := db.DB.Where("level_id = ?", levelID).First(&tpl).Error; err != nil {
		// Fallback: find template for the lowest level
		db.DB.Joins("JOIN member_levels ml ON ml.id = member_certificate_templates.level_id").
			Order("ml.level ASC").
			First(&tpl)
	}
	if tpl.ID > 0 {
		updates["cert_template_id"] = tpl.ID
	}

	db.DB.Model(&cert).Updates(updates)
}
