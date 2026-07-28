package service

import (
	"errors"
	"member/internal/models"
	"member/pkg/db"

	"gorm.io/gorm"
)

type MemberLevelService struct{}

// List returns all member levels ordered by level asc
func (s *MemberLevelService) List() ([]models.MemberLevel, error) {
	var list []models.MemberLevel
	if err := db.DB.Order("level ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// Get returns a single member level
func (s *MemberLevelService) Get(id uint64) (*models.MemberLevel, error) {
	var l models.MemberLevel
	if err := db.DB.First(&l, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("会员等级不存在")
		}
		return nil, err
	}
	return &l, nil
}

// Create creates a new member level
func (s *MemberLevelService) Create(req MemberLevelRequest) (*models.MemberLevel, error) {
	// Auto-assign next level number
	var max models.MemberLevel
	db.DB.Order("level DESC").First(&max)
	nextLevel := max.Level + 1

	l := models.MemberLevel{
		Name:        req.Name,
		Level:       nextLevel,
		Description: req.Description,
	}
	if err := db.DB.Create(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

// Update updates a member level
func (s *MemberLevelService) Update(id uint64, req MemberLevelRequest) error {
	return db.DB.Model(&models.MemberLevel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
	}).Error
}

// Delete deletes a member level
func (s *MemberLevelService) Delete(id uint64) error {
	// Check no members are using this level
	var count int64
	db.DB.Model(&models.Member{}).Where("member_level = ?",
		db.DB.Table("member_levels").Select("name").Where("id = ?", id)).Count(&count)
	// Actually get the level name first
	var l models.MemberLevel
	if err := db.DB.First(&l, id).Error; err != nil {
		return errors.New("会员等级不存在")
	}
	db.DB.Model(&models.Member{}).Where("member_level = ?", l.Name).Count(&count)
	if count > 0 {
		return errors.New("该等级下有会员，无法删除")
	}
	return db.DB.Delete(&models.MemberLevel{}, id).Error
}

// MoveUp decreases the level order (moves up in rank)
func (s *MemberLevelService) MoveUp(id uint64) error {
	var current models.MemberLevel
	if err := db.DB.First(&current, id).Error; err != nil {
		return errors.New("会员等级不存在")
	}
	if current.Level <= 0 {
		return errors.New("已是最高等级")
	}

	var above models.MemberLevel
	if err := db.DB.Where("level < ?", current.Level).Order("level DESC").First(&above).Error; err != nil {
		return errors.New("已是最高等级")
	}

	tx := db.DB.Begin()
	tx.Model(&current).Update("level", above.Level)
	tx.Model(&above).Update("level", current.Level)
	return tx.Commit().Error
}

// MoveDown increases the level order (moves down in rank)
func (s *MemberLevelService) MoveDown(id uint64) error {
	var current models.MemberLevel
	if err := db.DB.First(&current, id).Error; err != nil {
		return errors.New("会员等级不存在")
	}

	var below models.MemberLevel
	if err := db.DB.Where("level > ?", current.Level).Order("level ASC").First(&below).Error; err != nil {
		return errors.New("已是最低等级")
	}

	tx := db.DB.Begin()
	tx.Model(&current).Update("level", below.Level)
	tx.Model(&below).Update("level", current.Level)
	return tx.Commit().Error
}

type MemberLevelRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}
