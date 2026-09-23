package service

import (
	"errors"
	"fmt"
	"member/internal/models"
	"member/pkg/db"
	"time"

	"gorm.io/gorm"
)

// feeStandardMinYear 会费标准允许的最小年度（避免写入 1923 之类的脏数据）
const feeStandardMinYear = 2000

// validateFeeStandard 校验年度与金额边界（Upsert / BatchUpsert 共用）
func validateFeeStandard(year int, amount float64) error {
	if year < feeStandardMinYear || year > time.Now().Year()+10 {
		return fmt.Errorf("缴费年度不合法（应在 %d ~ 当前年份+10 之间）", feeStandardMinYear)
	}
	if amount < 0 {
		return errors.New("会费金额不能为负数")
	}
	return nil
}

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
	if err := validateFeeStandard(year, amount); err != nil {
		return nil, err
	}
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

// BatchUpsert creates or updates multiple fee standards at once.
// 整批放在一个事务里：原先逐条调用 Upsert，中途失败会留下“半套”数据。
func (s *FeeStandardService) BatchUpsert(levelID uint64, items []FeeStandardItem) error {
	if len(items) == 0 {
		return errors.New("请至少提交一条会费标准")
	}
	for _, item := range items {
		if err := validateFeeStandard(item.Year, item.Amount); err != nil {
			return err
		}
	}
	return db.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			var fs models.MemberFeeStandard
			err := tx.Where("level_id = ? AND year = ?", levelID, item.Year).First(&fs).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(&models.MemberFeeStandard{
					LevelID: levelID,
					Year:    item.Year,
					Amount:  item.Amount,
				}).Error; err != nil {
					return err
				}
				continue
			}
			if err != nil {
				return err
			}
			if err := tx.Model(&fs).Update("amount", item.Amount).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Delete removes a fee standard by ID
func (s *FeeStandardService) Delete(id uint64) error {
	var fs models.MemberFeeStandard
	if err := db.DB.First(&fs, id).Error; err != nil {
		return errors.New("会费标准不存在")
	}
	// 被会费记录引用时不允许删除，避免悬空 fee_standard_id
	var used int64
	if err := db.DB.Model(&models.FeeRecord{}).Where("fee_standard_id = ?", id).Count(&used).Error; err != nil {
		return err
	}
	if used > 0 {
		return errors.New("该会费标准已被会费记录使用，无法删除")
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
