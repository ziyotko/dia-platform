package service

import (
	"errors"
	"fmt"

	"base/internal/models"
	"base/pkg/db"

	"gorm.io/gorm"
)

type AppService struct{}

// Create 创建应用。应用编码唯一（重名直接报错），删除是物理删除，
// 编码删除后可以直接重新使用。
func (s AppService) Create(a *models.App) error {
	if a.Code == "" {
		return errors.New("请填写应用编码")
	}
	if err := validateAppFields(a); err != nil {
		return err
	}

	var count int64
	if err := db.DB.Model(&models.App{}).Where("code = ?", a.Code).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("应用编码已存在")
	}
	if err := db.DB.Create(a).Error; err != nil {
		// 并发下两个请求可能同时通过预检查，唯一索引兜底
		if isDuplicateEntry(err) {
			return errors.New("应用编码已存在")
		}
		return err
	}
	return nil
}

func (s AppService) Update(a *models.App) error {
	if err := validateAppFields(a); err != nil {
		return err
	}
	if err := ensureRecordExists(db.DB.Model(&models.App{}).Where("id = ?", a.ID), "应用不存在"); err != nil {
		return err
	}
	return db.DB.Model(a).Updates(map[string]interface{}{
		"name":         a.Name,
		"icon":         a.Icon,
		"type":         a.Type,
		"frontend_url": a.FrontendURL,
		"backend_url":  a.BackendURL,
		"api_prefix":   a.ApiPrefix,
		"status":       a.Status,
		"sort":         a.Sort,
		"description":  a.Description,
	}).Error
}

// Delete 删除应用。仍有租户开通该应用时拒绝删除，避免产生无归属的应用实例。
func (s AppService) Delete(id uint64) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.App{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errors.New("应用不存在")
		}
		var instances int64
		if err := tx.Model(&models.AppInstance{}).Where("app_id = ?", id).Count(&instances).Error; err != nil {
			return err
		}
		if instances > 0 {
			return fmt.Errorf("已有 %d 个租户开通了该应用，请先在「应用实例」中删除后再删除应用", instances)
		}
		return tx.Where("id = ?", id).Delete(&models.App{}).Error
	})
}

func (s AppService) GetByID(id uint64) (*models.App, error) {
	var a models.App
	err := db.DB.First(&a, id).Error
	return &a, err
}

func (s AppService) List(page, size int, keyword string) ([]models.App, int64, error) {
	var list []models.App
	var total int64
	query := db.DB.Model(&models.App{})
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Order("sort ASC, created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (s AppService) ListAllActive() ([]models.App, error) {
	var list []models.App
	err := db.DB.Where("status = ?", 1).Order("sort ASC").Find(&list).Error
	return list, err
}

// validateAppFields 校验应用各字段长度（对应 base_app 的定长列）。
func validateAppFields(a *models.App) error {
	return validateLengths(
		fieldLen{"应用编码", a.Code, 64},
		fieldLen{"应用名称", a.Name, 128},
		fieldLen{"图标", a.Icon, 256},
		fieldLen{"接入类型", a.Type, 32},
		fieldLen{"前端入口地址", a.FrontendURL, 512},
		fieldLen{"后端入口地址", a.BackendURL, 512},
		fieldLen{"API 前缀", a.ApiPrefix, 128},
		fieldLen{"描述", a.Description, 512},
	)
}
