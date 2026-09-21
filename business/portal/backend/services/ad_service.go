package services

import (
	"server/models"
	"server/utils"
)

type AdService struct{}

func (s *AdService) GetAds(name string, templateID int, columnID int, status int, page int, pageSize int) ([]models.Ad, int64, error) {
	var ads []models.Ad
	var total int64
	query := utils.DB.Model(&models.Ad{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if templateID > 0 {
		query = query.Where("template_id = ?", templateID)
	}
	if columnID > 0 {
		query = query.Where("column_id = ?", columnID)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err = query.Order("sort ASC, id DESC").Limit(pageSize).Offset(offset).Find(&ads).Error
	return ads, total, err
}

func (s *AdService) GetAdByID(id uint) (*models.Ad, error) {
	var ad models.Ad
	err := utils.DB.First(&ad, id).Error
	if err != nil {
		return nil, err
	}
	return &ad, nil
}

func (s *AdService) CreateAd(ad *models.Ad) error {
	return utils.DB.Create(ad).Error
}

func (s *AdService) UpdateAd(id uint, ad *models.Ad) error {
	var old models.Ad
	if err := utils.DB.First(&old, id).Error; err != nil {
		return err
	}
	updates := map[string]any{
		"name":        ad.Name,
		"template_id": ad.TemplateID,
		"column_id":   ad.ColumnID,
		"image":       ad.Image,
		"link":        ad.Link,
		"sort":        ad.Sort,
		"status":      ad.Status,
		"start_time":  ad.StartTime,
		"end_time":    ad.EndTime,
	}
	return utils.DB.Model(&old).Updates(updates).Error
}

func (s *AdService) UpdateAdStatus(id uint, status int) error {
	return utils.DB.Model(&models.Ad{}).Where("id = ?", id).Update("status", status).Error
}

func (s *AdService) DeleteAd(id uint) error {
	var ad models.Ad
	if err := utils.DB.First(&ad, id).Error; err != nil {
		return err
	}
	return utils.DB.Unscoped().Delete(&ad).Error
}
