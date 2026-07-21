package service

import (
	"base/internal/models"
	"base/pkg/db"
)

type AppService struct{}

func (s AppService) Create(a *models.App) error {
	return db.DB.Create(a).Error
}

func (s AppService) Update(a *models.App) error {
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

func (s AppService) Delete(id uint64) error {
	return db.DB.Delete(&models.App{BaseModel: models.BaseModel{ID: id}}).Error
}

func (s AppService) GetByID(id uint64) (*models.App, error) {
	var a models.App
	err := db.DB.First(&a, id).Error
	return &a, err
}

func (s AppService) GetByCode(code string) (*models.App, error) {
	var a models.App
	err := db.DB.Where("code = ?", code).First(&a).Error
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
