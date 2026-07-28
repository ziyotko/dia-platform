package service

import (
	"errors"
	"member/internal/models"
	"member/pkg/db"

	"gorm.io/gorm"
)

type MemberService struct{}

// GetMember returns a member by ID
func (s *MemberService) GetMember(id uint64) (*models.Member, error) {
	var m models.Member
	if err := db.DB.First(&m, id).Error; err != nil {
		return nil, errors.New("会员不存在")
	}
	return &m, nil
}

// ListMembers returns paginated member list (admin)
func (s *MemberService) ListMembers(page, size int, keyword, status, memberType string) ([]models.Member, int64, error) {
	var members []models.Member
	var total int64

	query := db.DB.Model(&models.Member{})
	if keyword != "" {
		kw := "%" + keyword + "%"
		query = query.Where("username LIKE ? OR company_name LIKE ? OR mobile LIKE ? OR email LIKE ?",
			kw, kw, kw, kw)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if memberType != "" {
		query = query.Where("member_type = ?", memberType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&members).Error; err != nil {
		return nil, 0, err
	}
	return members, total, nil
}

// UpdateMemberStatus updates a member's status (admin)
func (s *MemberService) UpdateMemberStatus(id uint64, status string) error {
	validStatuses := map[string]bool{
		models.MemberStatusRegistering:   true,
		models.MemberStatusPendingReview: true,
		models.MemberStatusPendingCert:   true,
		models.MemberStatusPendingPay:    true,
		models.MemberStatusActive:        true,
		models.MemberStatusRejected:      true,
		models.MemberStatusExpired:       true,
	}
	if !validStatuses[status] {
		return errors.New("无效的状态值")
	}
	return db.DB.Model(&models.Member{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateMemberLevel updates a member's level (admin)
func (s *MemberService) UpdateMemberLevel(id uint64, level string) error {
	return db.DB.Model(&models.Member{}).Where("id = ?", id).Update("member_level", level).Error
}

// DeleteMember soft-deletes a member (admin)
func (s *MemberService) DeleteMember(id uint64) error {
	var m models.Member
	if err := db.DB.First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("会员不存在")
		}
		return err
	}
	return db.DB.Delete(&m).Error
}
