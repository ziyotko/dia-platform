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

// GetCertificate returns a certificate by ID
func (s *CertificateService) GetCertificate(id uint64) (*models.Certificate, error) {
	var cert models.Certificate
	if err := db.DB.Preload("Member").First(&cert, id).Error; err != nil {
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
		Status:   "active",
	}
	if err := db.DB.Create(&cert).Error; err != nil {
		return nil, err
	}
	return &cert, nil
}

// UpdateCertificate updates a certificate (admin)
func (s *CertificateService) UpdateCertificate(id uint64, filePath, status string) error {
	updates := map[string]interface{}{}
	if filePath != "" {
		updates["file_path"] = filePath
	}
	if status != "" {
		updates["status"] = status
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
		ExpireAt: &models.LocalTime{Time: now.AddDate(1, 0, 0)},
		Status:   "active",
	}
	if err := db.DB.Create(&cert).Error; err != nil {
		return nil, err
	}
	return &cert, nil
}

// RenewMyCertificate marks old certificates as expired and creates a new one (member self-service)
func (s *CertificateService) RenewMyCertificate(memberID uint64) (*models.Certificate, error) {
	// Expire all active certificates for this member
	db.DB.Model(&models.Certificate{}).
		Where("member_id = ? AND status = 'active'", memberID).
		Update("status", "expired")

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
