package service

import (
	"strings"

	"base/internal/models"
	"base/pkg/db"

	"gorm.io/gorm"
)

// RenderMessageTemplate 用 vars 替换模板中的 {{key}} 占位符。
// 未提供取值的占位符保持原样，便于发送前人工补齐。
func RenderMessageTemplate(text string, vars map[string]string) string {
	if text == "" || len(vars) == 0 {
		return text
	}
	out := text
	for k, v := range vars {
		out = strings.ReplaceAll(out, "{{"+k+"}}", v)
		out = strings.ReplaceAll(out, "{{ "+k+" }}", v)
	}
	return out
}

type MessageTemplateService struct{}

func (s MessageTemplateService) Create(t *models.MessageTemplate) error {
	if err := validateTemplateFields(t); err != nil {
		return err
	}
	return db.DB.Create(t).Error
}

// validateTemplateFields 校验消息模板字段长度（对应 base_message_template 的定长列）。
func validateTemplateFields(t *models.MessageTemplate) error {
	return validateLengths(
		fieldLen{"模板编码", t.Code, 64},
		fieldLen{"模板名称", t.Name, 128},
		fieldLen{"渠道", t.Channel, 32},
		fieldLen{"主题", t.Subject, 256},
		fieldLen{"描述", t.Description, 512},
	)
}

func (s MessageTemplateService) Update(t *models.MessageTemplate, tenantID uint64) error {
	if err := validateTemplateFields(t); err != nil {
		return err
	}
	check := db.DB.Model(&models.MessageTemplate{}).Where("id = ?", t.ID)
	db := db.DB.Model(t).Where("id = ?", t.ID)
	if tenantID > 0 {
		check = check.Where("tenant_id = ?", tenantID)
		db = db.Where("tenant_id = ?", tenantID)
	}
	if err := ensureRecordExists(check, msgNotOwnedOrMissing); err != nil {
		return err
	}
	return db.Updates(map[string]interface{}{
		"name":        t.Name,
		"channel":     t.Channel,
		"subject":     t.Subject,
		"content":     t.Content,
		"variables":   t.Variables,
		"status":      t.Status,
		"description": t.Description,
	}).Error
}

func (s MessageTemplateService) Delete(id uint64, tenantID uint64) error {
	db := db.DB.Where("id = ?", id)
	if tenantID > 0 {
		db = db.Where("tenant_id = ?", tenantID)
	}
	return ensureDeleteAffected(db.Delete(&models.MessageTemplate{}), msgNotOwnedOrMissingDelete)
}

// GetByCode 按编码取启用中的模板：租户自定义模板优先，无则回退平台内置模板。
func (s MessageTemplateService) GetByCode(code string, tenantID uint64) (*models.MessageTemplate, error) {
	var list []models.MessageTemplate
	query := db.DB.Where("code = ? AND status = ?", code, 1)
	if tenantID > 0 {
		query = query.Where("tenant_id = ? OR tenant_id = 0", tenantID).Order("tenant_id DESC")
	}
	if err := query.Limit(1).Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &list[0], nil
}

func (s MessageTemplateService) GetByID(id uint64, tenantID uint64) (*models.MessageTemplate, error) {
	var t models.MessageTemplate
	query := db.DB.Where("id = ?", id)
	// 读取口径：本租户 + 平台内置（tenant_id = 0）均可读；写操作仍限本租户
	if tenantID > 0 {
		query = query.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}
	err := query.First(&t).Error
	return &t, err
}

func (s MessageTemplateService) List(page, size int, keyword string, tenantID uint64) ([]models.MessageTemplate, int64, error) {
	var list []models.MessageTemplate
	var total int64
	query := db.DB.Model(&models.MessageTemplate{})
	if tenantID > 0 {
		// 读取口径：本租户 + 平台内置（tenant_id = 0）
		query = query.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}
