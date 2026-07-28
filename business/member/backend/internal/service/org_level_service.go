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
	// Delete existing associations
	if err := db.DB.Where("org_id = ?", orgID).Delete(&models.MemberOrgLevel{}).Error; err != nil {
		return err
	}
	// Insert new associations
	for _, lid := range levelIDs {
		if err := db.DB.Create(&models.MemberOrgLevel{OrgID: orgID, LevelID: lid}).Error; err != nil {
			return err
		}
	}
	return nil
}
