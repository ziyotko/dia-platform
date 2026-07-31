package service

import (
	"errors"
	"member/internal/models"
	"member/pkg/db"

	"gorm.io/gorm"
)

type SystemConfigService struct{}

// List returns all system configs ordered by id
func (s *SystemConfigService) List() ([]models.SystemConfig, error) {
	var list []models.SystemConfig
	if err := db.DB.Order("id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// GetByKey returns a single config by key
func (s *SystemConfigService) GetByKey(key string) (*models.SystemConfig, error) {
	var cfg models.SystemConfig
	if err := db.DB.Where("`key` = ?", key).First(&cfg).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Create creates a new system config
func (s *SystemConfigService) Create(req SystemConfigRequest) (*models.SystemConfig, error) {
	if req.Key == "" {
		return nil, errors.New("配置项名称不能为空")
	}
	var count int64
	if err := db.DB.Model(&models.SystemConfig{}).Where("`key` = ?", req.Key).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("配置项已存在")
	}
	cfg := models.SystemConfig{
		Key:         req.Key,
		Value:       req.Value,
		Description: req.Description,
	}
	if err := db.DB.Create(&cfg).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Update updates an existing system config
func (s *SystemConfigService) Update(id uint64, req SystemConfigRequest) error {
	var cfg models.SystemConfig
	if err := db.DB.First(&cfg, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("配置项不存在")
		}
		return err
	}
	if req.Key != "" && req.Key != cfg.Key {
		var count int64
		if err := db.DB.Model(&models.SystemConfig{}).Where("`key` = ? AND id <> ?", req.Key, id).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("配置项名称已存在")
		}
	}
	updates := map[string]interface{}{}
	if req.Key != "" {
		updates["key"] = req.Key
	}
	if req.Value != "" {
		updates["value"] = req.Value
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if len(updates) == 0 {
		return nil
	}
	return db.DB.Model(&models.SystemConfig{}).Where("id = ?", id).Updates(updates).Error
}

// Delete removes a system config
func (s *SystemConfigService) Delete(id uint64) error {
	return db.DB.Delete(&models.SystemConfig{}, id).Error
}

// SystemConfigRequest is the create/update payload for system configs
type SystemConfigRequest struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
}
