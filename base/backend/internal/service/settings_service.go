package service

import (
	"base/internal/models"
	"base/pkg/db"

	"gorm.io/gorm"
)

type SettingsService struct{}

func (s SettingsService) GetAll() (map[string]map[string]string, error) {
	var list []models.Setting
	if err := db.DB.Find(&list).Error; err != nil {
		return nil, err
	}
	result := make(map[string]map[string]string)
	for _, item := range list {
		if result[item.Category] == nil {
			result[item.Category] = make(map[string]string)
		}
		result[item.Category][item.Key] = item.Value
	}
	return result, nil
}

func (s SettingsService) GetByCategory(category string) (map[string]string, error) {
	var list []models.Setting
	if err := db.DB.Where("category = ?", category).Find(&list).Error; err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, item := range list {
		result[item.Key] = item.Value
	}
	return result, nil
}

func (s SettingsService) BatchSave(items []models.Setting) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if item.Category == "" || item.Key == "" {
				continue
			}
			var existing models.Setting
			err := tx.Where("category = ? AND key = ?", item.Category, item.Key).First(&existing).Error
			if err == nil {
				existing.Value = item.Value
				existing.Type = item.Type
				existing.Remark = item.Remark
				if err := tx.Save(&existing).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Create(&item).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
