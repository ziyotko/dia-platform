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
	// Check if member already has a pending/approved application
	var existing models.Application
	if err := db.DB.Where("member_id = ? AND status NOT IN (?, ?)",
		memberID, models.AppStatusRejected, models.AppStatusDraft).First(&existing).Error; err == nil {
		return nil, errors.New("您已有进行中的入会申请")
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
	db.DB.Model(&models.Member{}).Where("id = ?", memberID).
		Updates(map[string]interface{}{
			"status":         models.MemberStatusPendingReview,
			"company_name":   req.CompanyName,
			"credit_code":    req.CreditCode,
			"legal_person":   req.LegalPerson,
			"contact_person": req.ContactPerson,
			"address":        req.Address,
			"member_type":    req.MemberType,
		})

	return &app, nil
}

// SaveDraft saves an application as draft
func (s *ApplicationService) SaveDraft(memberID uint64, req CreateAppRequest) (*models.Application, error) {
	app := models.Application{
		MemberID:   memberID,
		OrgID:      req.OrgID,
		Status:     models.AppStatusDraft,
		FormData:   req.FormData,
		SignedFile: req.SignedFile,
	}
	if err := db.DB.Create(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

// WithdrawApplication withdraws a pending application (sets back to draft)
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

	if err := db.DB.Model(&app).Update("status", models.AppStatusDraft).Error; err != nil {
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

// GetApplication returns an application by ID
func (s *ApplicationService) GetApplication(id uint64) (*models.Application, error) {
	var app models.Application
	if err := db.DB.Preload("Org").Preload("Member").First(&app, id).Error; err != nil {
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
		newMemberStatus = models.MemberStatusPendingCert
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

	// Update member status
	if err := tx.Model(&models.Member{}).Where("id = ?", app.MemberID).
		Update("status", newMemberStatus).Error; err != nil {
		tx.Rollback()
		return err
	}

	// If approved, auto-create certificate record and first-year fee
	if approved {
		now := time.Now()
		cert := models.Certificate{
			MemberID: app.MemberID,
			CertNo:   generateCertNo(app.MemberID),
			IssuedAt: &models.LocalTime{Time: now},
			ExpireAt: &models.LocalTime{Time: now.AddDate(1, 0, 0)},
			Status:   "active",
		}
		if err := tx.Create(&cert).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Create first-year fee record (using default amount)
		fee := models.FeeRecord{
			MemberID: app.MemberID,
			Year:     now.Year(),
			Amount:   2000.00, // Default, can be configured
			Status:   models.FeeStatusUnpaid,
		}
		if err := tx.Create(&fee).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
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
	return "CEEIA-" + time.Now().Format("2006") + "-" + padLeft(memberID)
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
