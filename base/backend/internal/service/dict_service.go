package service

import (
	"base/internal/models"
	"base/pkg/db"

	"gorm.io/gorm"
)

type DictService struct{}

type DictListQuery struct {
	Code   string
	Name   string
	Status int
	Page   int
	Size   int
}

func (s DictService) Create(d *models.Dict) error {
	return db.DB.Create(d).Error
}

func (s DictService) Update(d *models.Dict) error {
	return db.DB.Model(d).Updates(map[string]interface{}{
		"code":        d.Code,
		"name":        d.Name,
		"description": d.Description,
		"status":      d.Status,
	}).Error
}

func (s DictService) Delete(id uint64) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("dict_id = ?", id).Delete(&models.DictItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Dict{}, id).Error
	})
}

func (s DictService) GetByID(id uint64) (*models.Dict, error) {
	var d models.Dict
	err := db.DB.Preload("Items").First(&d, id).Error
	return &d, err
}

func (s DictService) List(q DictListQuery) ([]models.Dict, int64, error) {
	var list []models.Dict
	var total int64
	query := db.DB.Model(&models.Dict{})
	if q.Code != "" {
		query = query.Where("code LIKE ?", "%"+q.Code+"%")
	}
	if q.Name != "" {
		query = query.Where("name LIKE ?", "%"+q.Name+"%")
	}
	if q.Status >= 0 {
		query = query.Where("status = ?", q.Status)
	}
	query.Count(&total)
	offset := (q.Page - 1) * q.Size
	err := query.Order("created_at DESC").Offset(offset).Limit(q.Size).Find(&list).Error
	return list, total, err
}

func (s DictService) GetByCode(code string) (*models.Dict, error) {
	var d models.Dict
	err := db.DB.Where("code = ? AND status = ?", code, 1).Preload("Items", "status = ?", 1).First(&d).Error
	return &d, err
}

func (s DictService) SaveItems(dictID uint64, items []models.DictItem) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("dict_id = ?", dictID).Delete(&models.DictItem{}).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		for i := range items {
			items[i].DictID = dictID
			items[i].ID = 0
		}
		return tx.Create(&items).Error
	})
}
