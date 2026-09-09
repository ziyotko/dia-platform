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
	// 不显示管理员
	query = query.Where("is_admin = ?", false)
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

// UpdateMemberLevel updates a member's level (admin). Only active members may change
// level, and the target level must come from the member's paid memberships (会籍).
func (s *MemberService) UpdateMemberLevel(id uint64, level string) error {
	var m models.Member
	if err := db.DB.First(&m, id).Error; err != nil {
		return errors.New("会员不存在")
	}
	if m.Status != models.MemberStatusActive {
		return errors.New("仅正式会员可变更等级")
	}
	if level == "" {
		return errors.New("请选择会员等级")
	}

	// 目标等级必须来自该会员已缴费加入的会籍
	var cnt int64
	if err := db.DB.Model(&models.FeeRecord{}).
		Where("member_id = ? AND status = ?", id, models.FeeStatusPaid).
		Where("level_name = ? OR level_id IN (SELECT id FROM member_levels WHERE name = ?)", level, level).
		Count(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		return errors.New("该等级不在该会员已缴费加入的会籍中")
	}

	if err := db.DB.Model(&models.Member{}).Where("id = ?", id).Update("member_level", level).Error; err != nil {
		return err
	}

	// 同步更新该会员生效证书的等级
	var lvl models.MemberLevel
	if err := db.DB.Where("name = ?", level).First(&lvl).Error; err == nil {
		db.DB.Model(&models.Certificate{}).
			Where("member_id = ? AND status = ?", id, "active").
			Updates(map[string]interface{}{"level_id": lvl.ID, "level_name": lvl.Name})
	}

	return nil
}

// GetMemberAvailableLevels returns the levels this member has paid to join (会籍),
// used to populate the change-level dropdown options.
func (s *MemberService) GetMemberAvailableLevels(memberID uint64) ([]models.MemberLevel, error) {
	var levels []models.MemberLevel
	err := db.DB.Raw(`
		SELECT DISTINCT ml.id, ml.name, ml.level, ml.description, ml.created_at, ml.updated_at, ml.deleted_at
		FROM member_levels ml
		JOIN member_fee_records fr ON fr.level_id = ml.id
		WHERE fr.member_id = ? AND fr.status = ?
		ORDER BY ml.level ASC
	`, memberID, models.FeeStatusPaid).Scan(&levels).Error
	if err != nil {
		return nil, err
	}
	return levels, nil
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
