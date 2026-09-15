package service

import (
	"errors"
	"fmt"

	"base/internal/models"
	"base/pkg/db"

	"gorm.io/gorm"
)

type AppService struct{}

// Create 创建应用。
// 同编码的软删除记录会被恢复并覆盖字段（原因同租户：软删除不会释放唯一索引）。
func (s AppService) Create(a *models.App) error {
	if a.Code == "" {
		return errors.New("请填写应用编码")
	}

	var existing models.App
	err := db.DB.Unscoped().Where("code = ?", a.Code).First(&existing).Error
	switch {
	case err == nil:
		if !existing.DeletedAt.Valid {
			return errors.New("应用编码已存在")
		}
		if err := db.DB.Unscoped().Model(&models.App{}).Where("id = ?", existing.ID).Updates(map[string]interface{}{
			"deleted_at":   nil,
			"name":         a.Name,
			"icon":         a.Icon,
			"type":         a.Type,
			"frontend_url": a.FrontendURL,
			"backend_url":  a.BackendURL,
			"api_prefix":   a.ApiPrefix,
			"status":       a.Status,
			"sort":         a.Sort,
			"description":  a.Description,
		}).Error; err != nil {
			return err
		}
		return db.DB.Where("id = ?", existing.ID).First(a).Error
	case errors.Is(err, gorm.ErrRecordNotFound):
		return db.DB.Create(a).Error
	default:
		return err
	}
}

func (s AppService) Update(a *models.App) error {
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
