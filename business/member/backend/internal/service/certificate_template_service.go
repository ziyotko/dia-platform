package service

import (
	"errors"
	"member/internal/models"
	"member/pkg/db"
	"strings"
)

type CertificateTemplateService struct{}

// List returns all certificate templates ordered by created_at DESC
func (s *CertificateTemplateService) List() ([]models.MemberCertificateTemplate, error) {
	var list []models.MemberCertificateTemplate
	if err := db.DB.Preload("Level").Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// Get returns a single template by ID
func (s *CertificateTemplateService) Get(id uint64) (*models.MemberCertificateTemplate, error) {
	var tpl models.MemberCertificateTemplate
	if err := db.DB.Preload("Level").First(&tpl, id).Error; err != nil {
		return nil, errors.New("证书样式不存在")
	}
	return &tpl, nil
}

// Create creates a new certificate template
func (s *CertificateTemplateService) Create(req CertTemplateRequest) (*models.MemberCertificateTemplate, error) {
	// Verify level exists
	var level models.MemberLevel
	if err := db.DB.First(&level, req.LevelID).Error; err != nil {
		return nil, errors.New("会员等级不存在")
	}

	// Check if template already exists for this level
	var count int64
	db.DB.Model(&models.MemberCertificateTemplate{}).Where("level_id = ?", req.LevelID).Count(&count)
	if count > 0 {
		return nil, errors.New("该等级已有证书样式，请直接编辑")
	}

	tpl := models.MemberCertificateTemplate{
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
func (s *CertificateTemplateService) Update(id uint64, req UpdateCertTemplateRequest) error {
	var tpl models.MemberCertificateTemplate
	if err := db.DB.First(&tpl, id).Error; err != nil {
		return errors.New("证书样式不存在")
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return errors.New("请输入样式名称")
		}
		updates["name"] = name
	}
	if req.LevelID != nil {
		var level models.MemberLevel
		if err := db.DB.First(&level, *req.LevelID).Error; err != nil {
			return errors.New("会员等级不存在")
		}
		updates["level_id"] = *req.LevelID
	}
	if req.TemplateFile != nil {
		updates["template_file"] = *req.TemplateFile
	}
	if len(updates) == 0 {
		return nil
	}
	return db.DB.Model(&tpl).Updates(updates).Error
}

// Delete deletes a certificate template
func (s *CertificateTemplateService) Delete(id uint64) error {
	var tpl models.MemberCertificateTemplate
	if err := db.DB.First(&tpl, id).Error; err != nil {
		return errors.New("证书样式不存在")
	}
	return db.DB.Delete(&tpl).Error
}

// CertTemplateRequest is the request body for creating a template
type CertTemplateRequest struct {
	Name         string `json:"name" binding:"required"`
	LevelID      uint64 `json:"level_id" binding:"required"`
	TemplateFile string `json:"template_file"`
}

// UpdateCertTemplateRequest 证书样式更新请求。
// 字段使用指针以区分“未提供”（nil）与“清空”（指向空字符串）。
type UpdateCertTemplateRequest struct {
	Name         *string `json:"name"`
	LevelID      *uint64 `json:"level_id"`
	TemplateFile *string `json:"template_file"`
}
