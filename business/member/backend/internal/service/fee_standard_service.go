package service

import (
	"errors"
	"member/internal/models"
	"member/pkg/db"

	"gorm.io/gorm"
)

type FeeStandardService struct{}

// ListByLevel returns all fee standards for a given level, ordered by year desc
func (s *FeeStandardService) ListByLevel(levelID uint64) ([]models.MemberFeeStandard, error) {
	var list []models.MemberFeeStandard
	if err := db.DB.Where("level_id = ?", levelID).Order("year DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ListAll returns all fee standards grouped by level, ordered by year desc
func (s *FeeStandardService) ListAll() ([]models.MemberFeeStandard, error) {
	var list []models.MemberFeeStandard
	if err := db.DB.Preload("Level").Order("level_id ASC, year DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// GetByLevelAndYear retrieves a single fee standard
func (s *FeeStandardService) GetByLevelAndYear(levelID uint64, year int) (*models.MemberFeeStandard, error) {
	var fs models.MemberFeeStandard
	if err := db.DB.Where("level_id = ? AND year = ?", levelID, year).First(&fs).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("该等级对应年份的会费标准未设置")
		}
		return nil, err
	}
	return &fs, nil
}

// Get returns a fee standard by ID
func (s *FeeStandardService) Get(id uint64) (*models.MemberFeeStandard, error) {
	var fs models.MemberFeeStandard
	if err := db.DB.Preload("Level").First(&fs, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("会费标准不存在")
		}
		return nil, err
	}
	return &fs, nil
}

// Upsert creates or updates a fee standard for a given level + year
func (s *FeeStandardService) Upsert(levelID uint64, year int, amount float64) (*models.MemberFeeStandard, error) {
	// Verify the level exists
	var level models.MemberLevel
	if err := db.DB.First(&level, levelID).Error; err != nil {
		return nil, errors.New("会员等级不存在")
	}

	var fs models.MemberFeeStandard
	result := db.DB.Where("level_id = ? AND year = ?", levelID, year).First(&fs)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Create
		fs = models.MemberFeeStandard{
			LevelID: levelID,
			Year:    year,
			Amount:  amount,
		}
		if err := db.DB.Create(&fs).Error; err != nil {
			return nil, err
		}
	} else if result.Error != nil {
		return nil, result.Error
	} else {
		// Update
		if err := db.DB.Model(&fs).Update("amount", amount).Error; err != nil {
			return nil, err
		}
	}
	// Reload with Level
	db.DB.Preload("Level").First(&fs, fs.ID)
	return &fs, nil
}

// BatchUpsert creates or updates multiple fee standards at once
func (s *FeeStandardService) BatchUpsert(levelID uint64, items []FeeStandardItem) error {
	for _, item := range items {
		if _, err := s.Upsert(levelID, item.Year, item.Amount); err != nil {
			return err
		}
	}
	return nil
}

// Delete removes a fee standard by ID
func (s *FeeStandardService) Delete(id uint64) error {
	var fs models.MemberFeeStandard
	if err := db.DB.First(&fs, id).Error; err != nil {
		return errors.New("会费标准不存在")
	}
	return db.DB.Delete(&fs).Error
}

// FeeStandardItem is a request item for batch upsert
type FeeStandardItem struct {
	Year   int     `json:"year"`
	Amount float64 `json:"amount"`
}

// FeeStandardUpsertRequest is the request body for batch upsert
type FeeStandardUpsertRequest struct {
	LevelID uint64            `json:"level_id" binding:"required"`
	Items   []FeeStandardItem `json:"items" binding:"required"`
}
