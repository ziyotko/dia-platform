package service

import (
	"member/internal/models"
	"member/pkg/db"
)

type OrgLevelService struct{}

// GetOrgLevels returns all level IDs associated with an organization
func (s *OrgLevelService) GetOrgLevels(orgID uint64) ([]models.MemberOrgLevel, error) {
	var levels []models.MemberOrgLevel
	if err := db.DB.Preload("Level").Where("org_id = ?", orgID).Find(&levels).Error; err != nil {
		return nil, err
	}
	return levels, nil
}

// SetOrgLevels replaces all level associations for an organization
func (s *OrgLevelService) SetOrgLevels(orgID uint64, levelIDs []uint64) error {
	tx := db.DB.Begin()

	// 物理删除旧关联：GORM 默认 Delete 为软删（写入 deleted_at），而唯一索引
	// idx_org_level 只含 (org_id, level_id)，软删记录仍占索引导致重新插入重复报错。
	if err := tx.Unscoped().Where("org_id = ?", orgID).Delete(&models.MemberOrgLevel{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 插入新关联（去重并跳过空值，避免重复键冲突）
	seen := map[uint64]bool{}
	for _, lid := range levelIDs {
		if lid == 0 || seen[lid] {
			continue
		}
		seen[lid] = true
		if err := tx.Create(&models.MemberOrgLevel{OrgID: orgID, LevelID: lid}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}
