package service

import (
	"errors"
	"member/internal/models"
	"member/pkg/db"
	"strconv"
	"strings"
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

// GetMemberFeeInfo returns a member's org and level info used when creating a fee record.
// 解析优先级：
//  1. 已通过的入会申请 —— 申请机构 + 该机构最低等级；
//  2. 已缴费（含免缴）费用记录 —— 费用机构（通常是总会）+ 费用等级，覆盖管理员直录会员；
//  3. member_user_orgs 加入记录 —— 分会/代表机构 + 其等级；
//  4. 仅 member_users.member_level（无机构信息）。
func (s *FeeService) GetMemberFeeInfo(memberID uint64) (orgID, levelID uint64, orgName, levelName string, err error) {
	var member models.Member
	if err := db.DB.First(&member, memberID).Error; err != nil {
		return 0, 0, "", "", errors.New("会员不存在")
	}
	// member_level 存等级 ID 字符串（兼容历史名称），作为各分支的等级兑底
	memberLevelID, memberLevelName := resolveMemberLevel(member.MemberLevel)

	// 1) 已通过的入会申请
	var app models.Application
	if err := db.DB.Preload("Org").Where("member_id = ? AND status = ?", memberID, models.AppStatusApproved).
		Order("created_at DESC").First(&app).Error; err == nil {
		orgID = app.OrgID
		orgName = app.Org.Name
		levelID, levelName = orgMinLevelInfo(app.OrgID)
		if levelID == 0 {
			levelID, levelName = memberLevelID, memberLevelName
		}
		return orgID, levelID, orgName, levelName, nil
	}

	// 2) 最近一条已缴费（含免缴）费用记录
	var fee models.FeeRecord
	if err := db.DB.Where("member_id = ? AND status = ? AND org_name <> ''", memberID, models.FeeStatusPaid).
		Order("year DESC, id DESC").First(&fee).Error; err == nil {
		orgID, orgName = fee.OrgID, fee.OrgName
		levelID, levelName = fee.LevelID, fee.LevelName
		if levelID == 0 {
			if lvID, lvName := orgMinLevelInfo(orgID); lvID > 0 {
				levelID, levelName = lvID, lvName
			} else {
				levelID, levelName = memberLevelID, memberLevelName
			}
		}
		return orgID, levelID, orgName, levelName, nil
	}

	// 3) 会员加入的组织（member_user_orgs，分会/代表机构）
	var mo models.MemberOrganization
	if err := db.DB.Preload("Org").Where("member_id = ?", memberID).
		Order("joined_at DESC, id DESC").First(&mo).Error; err == nil {
		orgID = mo.OrgID
		orgName = mo.Org.Name
		if mo.LevelID > 0 {
			levelID = mo.LevelID
			levelName = memberLevelNameByID(mo.LevelID)
		}
		if levelID == 0 {
			if lvID, lvName := orgMinLevelInfo(orgID); lvID > 0 {
				levelID, levelName = lvID, lvName
			} else {
				levelID, levelName = memberLevelID, memberLevelName
			}
		}
		return orgID, levelID, orgName, levelName, nil
	}

	// 4) 仅有会员等级，无机构
	if memberLevelID > 0 || memberLevelName != "" {
		return 0, memberLevelID, "", memberLevelName, nil
	}

	return 0, 0, "", "", errors.New("未找到该会员的机构/等级信息，请先完善入会信息")
}

// orgMinLevelInfo 返回机构支持的最小等级（id + 名称），未配置时返回 (0, "")。
func orgMinLevelInfo(orgID uint64) (uint64, string) {
	if orgID == 0 {
		return 0, ""
	}
	ol, err := findMinOrgLevel(orgID)
	if err != nil || ol.LevelID == 0 {
		return 0, ""
	}
	return ol.LevelID, ol.Level.Name
}

// memberLevelNameByID 按等级 ID 取名称。
func memberLevelNameByID(levelID uint64) string {
	if levelID == 0 {
		return ""
	}
	var lvl models.MemberLevel
	if err := db.DB.First(&lvl, levelID).Error; err != nil {
		return ""
	}
	return lvl.Name
}

// resolveMemberLevel 兼容 member_users.member_level 既可能是等级 ID 字符串、也可能是历史等级名称。
func resolveMemberLevel(value string) (uint64, string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, ""
	}
	if id := parseUint(value); id > 0 {
		if name := memberLevelNameByID(id); name != "" {
			return id, name
		}
		return id, value
	}
	var lvl models.MemberLevel
	if err := db.DB.Where("name = ?", value).First(&lvl).Error; err == nil {
		return lvl.ID, lvl.Name
	}
	return 0, value
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
