package service

import (
	"errors"
	"member/internal/models"
	"member/pkg/db"
	"member/pkg/utils"
	"strings"
)

type CertificateTemplateService struct{}

// List returns all certificate templates ordered by created_at DESC
func (s *CertificateTemplateService) List() ([]models.MemberCertificateTemplate, error) {
	var list []models.MemberCertificateTemplate
	if err := db.DB.Preload("Level").Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	for i := range list {
		list[i].FileExists = templateFileExists(list[i].TemplateFile)
	}
	return list, nil
}

// Get returns a single template by ID
func (s *CertificateTemplateService) Get(id uint64) (*models.MemberCertificateTemplate, error) {
	var tpl models.MemberCertificateTemplate
	if err := db.DB.Preload("Level").First(&tpl, id).Error; err != nil {
		return nil, errors.New("证书样式不存在")
	}
	tpl.FileExists = templateFileExists(tpl.TemplateFile)
	return &tpl, nil
}

// templateFileExists 判断模板文件路径能否解析到服务器上的真实文件（空值不算存在）。
func templateFileExists(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	_, ok := utils.LocalUploadPath(path)
	return ok
}

// validateTemplateFile 校验上传后的模板文件路径真实可用，避免存入无法解析的路径。
func validateTemplateFile(path string) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	if _, ok := utils.LocalUploadPath(path); !ok {
		return errors.New("模板文件不存在或路径非法，请重新上传")
	}
	return nil
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

	if err := validateTemplateFile(req.TemplateFile); err != nil {
		return nil, err
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
		// 仅当模板文件发生变化时校验，避免历史脏路径堵住改名/改等级等无关编辑。
		if *req.TemplateFile != tpl.TemplateFile {
			if err := validateTemplateFile(*req.TemplateFile); err != nil {
				return err
			}
		}
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
