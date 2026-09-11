package service

import (
	"errors"
	"time"

	"member/internal/models"
	"member/pkg/db"
)

type MemberOrgService struct{}

// GetMyOrgs returns the member's joined organizations
func (s *MemberOrgService) GetMyOrgs(memberID uint64) ([]models.MemberOrganization, error) {
	var orgs []models.MemberOrganization
	if err := db.DB.Preload("Org").Where("member_id = ?", memberID).Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}

// JoinOrg joins an organization
func (s *MemberOrgService) JoinOrg(memberID, orgID, levelID uint64) error {
	// Check if org exists
	var org models.Organization
	if err := db.DB.First(&org, orgID).Error; err != nil {
		return errors.New("组织不存在")
	}
	// Check if already joined
	var exist models.MemberOrganization
	if err := db.DB.Where("member_id = ? AND org_id = ?", memberID, orgID).First(&exist).Error; err == nil {
		return errors.New("已加入该组织")
	}
	// 会员需有缴费加入的组织（已批准的入会申请，或已缴费/免缴的会费记录）
	var approvedCount int64
	db.DB.Model(&models.Application{}).Where("member_id = ? AND status = ?", memberID, models.AppStatusApproved).Count(&approvedCount)
	var paidFeeCount int64
	db.DB.Model(&models.FeeRecord{}).Where("member_id = ? AND status = ? AND org_name <> ''", memberID, models.FeeStatusPaid).Count(&paidFeeCount)
	if approvedCount == 0 && paidFeeCount == 0 {
		return errors.New("暂无缴费加入的组织，无法加入新组织")
	}
	// If levelID is provided, verify it belongs to this org
	if levelID > 0 {
		var orgLevel models.MemberOrgLevel
		if err := db.DB.Where("org_id = ? AND level_id = ?", orgID, levelID).First(&orgLevel).Error; err != nil {
			return errors.New("该组织不支持所选会员级别")
		}
	}
	mo := models.MemberOrganization{
		MemberID: memberID,
		OrgID:    orgID,
		LevelID:  levelID,
		JoinedAt: time.Now(),
	}
	return db.DB.Create(&mo).Error
}

// LeaveOrg leaves an organization
func (s *MemberOrgService) LeaveOrg(memberID, orgID uint64) error {
	if err := s.checkLeavable(orgID); err != nil {
		return err
	}
	return db.DB.Where("member_id = ? AND org_id = ?", memberID, orgID).Delete(&models.MemberOrganization{}).Error
}

// LeaveOrgByID leaves by membership record ID
func (s *MemberOrgService) LeaveOrgByID(memberID, id uint64) error {
	var mo models.MemberOrganization
	if err := db.DB.Where("id = ? AND member_id = ?", id, memberID).First(&mo).Error; err != nil {
		return errors.New("未找到该加入记录")
	}
	if err := s.checkLeavable(mo.OrgID); err != nil {
		return err
	}
	return db.DB.Where("id = ? AND member_id = ?", id, memberID).Delete(&models.MemberOrganization{}).Error
}

// checkLeavable prevents a member from leaving the root (level-1) organization
func (s *MemberOrgService) checkLeavable(orgID uint64) error {
	var org models.Organization
	if err := db.DB.First(&org, orgID).Error; err != nil {
		return errors.New("组织不存在")
	}
	if org.ParentID == 0 {
		return errors.New("总会为一级组织，不可自行退出")
	}
	return nil
}
