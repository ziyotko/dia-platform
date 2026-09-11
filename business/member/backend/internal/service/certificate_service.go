package service

import (
	"errors"
	"member/internal/models"
	"member/pkg/db"
	"time"
)

type CertificateService struct{}

// GetMyCertificates returns the member's certificates
func (s *CertificateService) GetMyCertificates(memberID uint64) ([]models.Certificate, error) {
	var certs []models.Certificate
	if err := db.DB.Where("member_id = ?", memberID).Order("created_at DESC").Find(&certs).Error; err != nil {
		return nil, err
	}
	return certs, nil
}

// GetCertificate returns a certificate by ID (owner only)
func (s *CertificateService) GetCertificate(id, memberID uint64) (*models.Certificate, error) {
	var cert models.Certificate
	if err := db.DB.Preload("Member").Where("id = ? AND member_id = ?", id, memberID).First(&cert).Error; err != nil {
		return nil, errors.New("证书不存在")
	}
	return &cert, nil
}

// CreateCertificate creates a certificate (admin)
func (s *CertificateService) CreateCertificate(req CreateCertRequest) (*models.Certificate, error) {
	cert := models.Certificate{
		MemberID: req.MemberID,
		CertNo:   req.CertNo,
		IssuedAt: req.IssuedAt,
		ExpireAt: req.ExpireAt,
		FilePath: req.FilePath,
		Status:   models.CertStatusActive,
	}
	if err := db.DB.Create(&cert).Error; err != nil {
		return nil, err
	}
	return &cert, nil
}

// UpdateCertificate updates a certificate (admin).
// filePath/status 使用指针区分“未提供”（nil）与“清空”（指向空字符串）。
func (s *CertificateService) UpdateCertificate(id uint64, filePath, status *string) error {
	updates := map[string]interface{}{}
	if filePath != nil {
		updates["file_path"] = *filePath
	}
	if status != nil {
		updates["status"] = *status
	}
	if len(updates) == 0 {
		return nil
	}
	return db.DB.Model(&models.Certificate{}).Where("id = ?", id).Updates(updates).Error
}

// GenerateCertificateForMember generates a certificate for a member (admin)
func (s *CertificateService) GenerateCertificateForMember(memberID uint64) (*models.Certificate, error) {
	now := time.Now()
	cert := models.Certificate{
		MemberID: memberID,
		CertNo:   "XXXXXX-" + now.Format("2006") + "-" + padLeftGen(memberID),
		IssuedAt: &models.LocalTime{Time: now},
		ExpireAt: &models.LocalTime{Time: time.Date(now.Year(), 12, 31, 23, 59, 59, 0, now.Location())},
		Status:   models.CertStatusActive,
	}
	if err := db.DB.Create(&cert).Error; err != nil {
		return nil, err
	}
	return &cert, nil
}

// RenewMyCertificate marks old certificates as expired and creates a new one (member self-service).
// 仅正式会员且当年已缴费方可续证，防止任意会员刷证。
func (s *CertificateService) RenewMyCertificate(memberID uint64) (*models.Certificate, error) {
	var member models.Member
	if err := db.DB.First(&member, memberID).Error; err != nil {
		return nil, errors.New("会员不存在")
	}
	if member.Status != models.MemberStatusActive {
		return nil, errors.New("仅正式会员可续证")
	}

	// 校验当年已缴费
	var paidCount int64
	if err := db.DB.Model(&models.FeeRecord{}).
		Where("member_id = ? AND year = ? AND status = ?", memberID, time.Now().Year(), models.FeeStatusPaid).
		Count(&paidCount).Error; err != nil {
		return nil, err
	}
	if paidCount == 0 {
		return nil, errors.New("当年尚未缴费，暂不能续证")
	}

	// Expire all active certificates for this member
	if err := db.DB.Model(&models.Certificate{}).
		Where("member_id = ? AND status = ?", memberID, models.CertStatusActive).
		Update("status", models.CertStatusExpired).Error; err != nil {
		return nil, err
	}

	// Create a new certificate
	return s.GenerateCertificateForMember(memberID)
}

type CreateCertRequest struct {
	MemberID uint64            `json:"member_id" binding:"required"`
	CertNo   string            `json:"cert_no" binding:"required"`
	IssuedAt *models.LocalTime `json:"issued_at"`
	ExpireAt *models.LocalTime `json:"expire_at"`
	FilePath string            `json:"file_path"`
}

func padLeftGen(id uint64) string {
	result := ""
	n := id
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	for len(result) < 6 {
		result = "0" + result
	}
	return result
}
