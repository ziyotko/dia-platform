package service

import (
	"errors"
	"member/internal/models"
	"member/pkg/db"
	"time"
)

type ApplicationService struct{}

// CreateApplication creates a new membership application
func (s *ApplicationService) CreateApplication(memberID uint64, req CreateAppRequest) (*models.Application, error) {
	if req.OrgID == 0 {
		return nil, errors.New("请选择入会机构")
	}

	// Check if member already has a pending/approved application
	var existing models.Application
	if err := db.DB.Where("member_id = ? AND status <> ?",
		memberID, models.AppStatusRejected).First(&existing).Error; err == nil {
		return nil, errors.New("您已有进行中的入会申请")
	}

	// 载入会员：正式会员（active）再次申请加入其它机构时不改变其会员状态，
	// 仅对尚未成为正式会员的账号推进为“待审核”，避免申请把正式会员降级。
	var member models.Member
	if err := db.DB.First(&member, memberID).Error; err != nil {
		return nil, errors.New("会员不存在")
	}

	app := models.Application{
		MemberID:   memberID,
		OrgID:      req.OrgID,
		Status:     models.AppStatusPendingReview,
		FormData:   req.FormData,
		SignedFile: req.SignedFile,
	}
	if err := db.DB.Create(&app).Error; err != nil {
		return nil, err
	}

	// Update member status to pending review
	// Only update fields actually provided; preserve member_type, legal_person, member_level
	updates := map[string]interface{}{}
	if member.Status != models.MemberStatusActive {
		updates["status"] = models.MemberStatusPendingReview
	}
	if req.CompanyName != "" {
		updates["company_name"] = req.CompanyName
	}
	if req.CreditCode != "" {
		updates["credit_code"] = req.CreditCode
	}
	if req.ContactPerson != "" {
		updates["contact_person"] = req.ContactPerson
	}
	if req.Address != "" {
		updates["address"] = req.Address
	}
	if len(updates) > 0 {
		db.DB.Model(&models.Member{}).Where("id = ?", memberID).Updates(updates)
	}

	return &app, nil
}

// WithdrawApplication withdraws a pending application (deletes the record)
func (s *ApplicationService) WithdrawApplication(id, memberID uint64) error {
	var app models.Application
	if err := db.DB.First(&app, id).Error; err != nil {
		return errors.New("申请不存在")
	}
	if app.MemberID != memberID {
		return errors.New("无权操作该申请")
	}
	if app.Status != models.AppStatusPendingReview {
		return errors.New("仅待审核状态的申请可以撤回")
	}

	// 撤回即删除该待审核申请（不再保留草稿状态）
	if err := db.DB.Delete(&models.Application{}, app.ID).Error; err != nil {
		return err
	}

	// Reset member status back to allow re-application
	db.DB.Model(&models.Member{}).Where("id = ?", memberID).
		Where("status = ?", models.MemberStatusPendingReview).
		Update("status", models.MemberStatusRegistering)

	return nil
}

// GetMyApplications returns the member's applications
func (s *ApplicationService) GetMyApplications(memberID uint64) ([]models.Application, error) {
	var apps []models.Application
	if err := db.DB.Preload("Org").Where("member_id = ?", memberID).
		Order("created_at DESC").Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

// GetApplication returns an application by ID (owner only)
func (s *ApplicationService) GetApplication(id, memberID uint64) (*models.Application, error) {
	var app models.Application
	if err := db.DB.Preload("Org").Preload("Member").Where("id = ? AND member_id = ?", id, memberID).First(&app).Error; err != nil {
		return nil, errors.New("申请不存在")
	}
	return &app, nil
}

// ReviewApplication reviews an application (admin)
func (s *ApplicationService) ReviewApplication(id, reviewerID uint64, approved bool, comment string) error {
	var app models.Application
	if err := db.DB.First(&app, id).Error; err != nil {
		return errors.New("申请不存在")
	}
	if app.Status != models.AppStatusPendingReview {
		return errors.New("该申请不在待审核状态")
	}

	newStatus := models.AppStatusRejected
	newMemberStatus := models.MemberStatusRejected
	if approved {
		newStatus = models.AppStatusApproved
		newMemberStatus = models.MemberStatusPendingPay
	}

	tx := db.DB.Begin()

	// Update application
	if err := tx.Model(&app).Updates(map[string]interface{}{
		"status":         newStatus,
		"review_comment": comment,
		"reviewer_id":    reviewerID,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update member status — 正式会员不因新的入会申请审批而改变状态
	var member models.Member
	if err := tx.First(&member, app.MemberID).Error; err != nil {
		tx.Rollback()
		return errors.New("会员不存在")
	}
	if member.Status != models.MemberStatusActive {
		if err := tx.Model(&models.Member{}).Where("id = ?", app.MemberID).
			Update("status", newMemberStatus).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 审批通过：会员自动加入总会（一级组织，parent_id = 0）。
	// 总会会籍以费用记录体现（会费标准按总会等级），
	// 分支机构/代表机构的关系写入 member_user_orgs。
	if approved {
		now := time.Now()

		// 定位申请机构及其所属总会
		var appliedOrg models.Organization
		if err := db.DB.First(&appliedOrg, app.OrgID).Error; err != nil {
			// 机构不存在时按申请机构本身处理（兼容历史数据）
			appliedOrg.ID = app.OrgID
		}
		rootOrg := appliedOrg
		for i := 0; i < 10 && rootOrg.ParentID != 0; i++ {
			var parent models.Organization
			if err := db.DB.First(&parent, rootOrg.ParentID).Error; err != nil {
				break
			}
			rootOrg = parent
		}

		// Look up the minimum member level for the 总会 and get its fee standard for current year;
		// 分会/代表机构沿用总会等级，总会未配置等级时回退到申请机构自身的配置。
		orgLevel, lvlErr := findMinOrgLevel(rootOrg.ID)
		if lvlErr != nil && appliedOrg.ID != rootOrg.ID {
			orgLevel, lvlErr = findMinOrgLevel(appliedOrg.ID)
		}

		var feeAmount float64 = 2000.00 // fallback default
		var feeStandardID, levelID uint64
		var levelName string
		if lvlErr == nil {
			// Found the minimum level for the 总会
			levelID = orgLevel.LevelID
			levelName = orgLevel.Level.Name

			var std models.MemberFeeStandard
			if err := db.DB.Where("level_id = ? AND year = ?", levelID, now.Year()).First(&std).Error; err == nil {
				feeAmount = std.Amount
				feeStandardID = std.ID
			}
		}

		// Look up certificate template for this level; fallback to lowest level's template
		var certTplID uint64
		var tpl models.MemberCertificateTemplate
		if err := db.DB.Where("level_id = ?", levelID).First(&tpl).Error; err != nil {
			db.DB.Joins("JOIN member_levels ml ON ml.id = member_certificate_templates.level_id").
				Order("ml.level ASC").
				First(&tpl)
		}
		if tpl.ID > 0 {
			certTplID = tpl.ID
		}

		cert := models.Certificate{
			MemberID:       app.MemberID,
			CertNo:         generateCertNo(app.MemberID),
			IssuedAt:       &models.LocalTime{Time: now},
			ExpireAt:       &models.LocalTime{Time: time.Date(now.Year(), 12, 31, 23, 59, 59, 0, now.Location())},
			Status:         models.CertStatusActive,
			LevelID:        levelID,
			LevelName:      levelName,
			CertTemplateID: certTplID,
		}
		if err := tx.Create(&cert).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 首年会费记录：机构为总会，会费标准按总会等级
		fee := models.FeeRecord{
			MemberID:      app.MemberID,
			Year:          now.Year(),
			Amount:        feeAmount,
			Status:        models.FeeStatusUnpaid,
			FeeStandardID: feeStandardID,
			LevelID:       levelID,
			LevelName:     levelName,
			OrgID:         rootOrg.ID,
			OrgName:       rootOrg.Name,
		}
		if err := tx.Create(&fee).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 申请的是分支机构/代表机构时，写入 member_user_orgs（总会不写入该表）
		if appliedOrg.ID != rootOrg.ID {
			var exists int64
			tx.Model(&models.MemberOrganization{}).
				Where("member_id = ? AND org_id = ?", app.MemberID, appliedOrg.ID).Count(&exists)
			if exists == 0 {
				mo := models.MemberOrganization{
					MemberID: app.MemberID,
					OrgID:    appliedOrg.ID,
					LevelID:  levelID,
					JoinedAt: now,
				}
				if err := tx.Create(&mo).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		}
	}

	return tx.Commit().Error
}

// findMinOrgLevel 返回机构支持的最小会员等级（按等级升序），未配置时返回错误。
func findMinOrgLevel(orgID uint64) (models.MemberOrgLevel, error) {
	var orgLevel models.MemberOrgLevel
	err := db.DB.Preload("Level").
		Joins("JOIN member_levels ml ON ml.id = member_org_levels.level_id").
		Where("member_org_levels.org_id = ?", orgID).
		Order("ml.level ASC").
		First(&orgLevel).Error
	return orgLevel, err
}

// ListApplications lists applications (admin)
func (s *ApplicationService) ListApplications(page, size int, status string) ([]models.Application, int64, error) {
	var apps []models.Application
	var total int64

	query := db.DB.Model(&models.Application{}).Preload("Member").Preload("Org")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&apps).Error; err != nil {
		return nil, 0, err
	}
	return apps, total, nil
}

type CreateAppRequest struct {
	OrgID         uint64 `json:"org_id"`
	FormData      string `json:"form_data"`
	SignedFile    string `json:"signed_file"`
	CompanyName   string `json:"company_name"`
	CreditCode    string `json:"credit_code"`
	LegalPerson   string `json:"legal_person"`
	ContactPerson string `json:"contact_person"`
	Address       string `json:"address"`
	MemberType    string `json:"member_type"`
}

func generateCertNo(memberID uint64) string {
	return "XXXXXX-" + time.Now().Format("2006") + "-" + padLeft(memberID)
}

func padLeft(id uint64) string {
	s := ""
	n := id
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	for len(s) < 6 {
		s = "0" + s
	}
	return s
}
