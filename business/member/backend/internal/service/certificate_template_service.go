package service
package service

import (
	"errors"
	"member/internal/models"
	"member/pkg/db"
)

type CertificateTemplateService struct{}

// List returns all certificate templates ordered by created_at DESC
func (s *CertificateTemplateService) List() ([]models.CertificateTemplate, error) {
	var list []models.CertificateTemplate
	if err := db.DB.Preload("Level").Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// Get returns a single template by ID
func (s *CertificateTemplateService) Get(id uint64) (*models.CertificateTemplate, error) {
	var tpl models.CertificateTemplate
	if err := db.DB.Preload("Level").First(&tpl, id).Error; err != nil {
		return nil, errors.New("证书样式不存在")
	}
	return &tpl, nil
}

// Create creates a new certificate template
func (s *CertificateTemplateService) Create(req CertTemplateRequest) (*models.CertificateTemplate, error) {
	// Verify level exists
	var level models.MemberLevel
	if err := db.DB.First(&level, req.LevelID).Error; err != nil {
		return nil, errors.New("会员等级不存在")
	}

	tpl := models.CertificateTemplate{
		Name:         req.Name,
		LevelID:      req.LevelID,
		TemplateFile: req.TemplateFile,
	}
	if err := db.DB.Create(&tpl).Error; err != nil {
		return nil, err
	}
	// Reload with level data
	db.DB.Preload("Level").First(&tpl, tpl.ID)
	return &tpl, nil
}

// Update updates a certificate template
func (s *CertificateTemplateService) Update(id uint64, req CertTemplateRequest) error {
	var tpl models.CertificateTemplate
	if err := db.DB.First(&tpl, id).Error; err != nil {
		return errors.New("证书样式不存在")
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.LevelID > 0 {
		var level models.MemberLevel
		if err := db.DB.First(&level, req.LevelID).Error; err != nil {
			return errors.New("会员等级不存在")
		}
		updates["level_id"] = req.LevelID
	}
	if req.TemplateFile != "" {
		updates["template_file"] = req.TemplateFile
	}
	return db.DB.Model(&tpl).Updates(updates).Error
}

// Delete deletes a certificate template
func (s *CertificateTemplateService) Delete(id uint64) error {
	var tpl models.CertificateTemplate
	if err := db.DB.First(&tpl, id).Error; err != nil {
		return errors.New("证书样式不存在")
	}
	return db.DB.Delete(&tpl).Error
}

// CertTemplateRequest is the request body for creating/updating a template
type CertTemplateRequest struct {
	Name         string `json:"name" binding:"required"`
	LevelID      uint64 `json:"level_id" binding:"required"`
	TemplateFile string `json:"template_file"`
}
