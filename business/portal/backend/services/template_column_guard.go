package services

import (
	"errors"

	"server/models"
	"server/utils"
)

var (
	// ErrTemplateRequired 未选择所属模板
	ErrTemplateRequired = errors.New("请选择所属模板")
	// ErrColumnRequired 未选择所属栏目
	ErrColumnRequired = errors.New("请选择所属栏目")
	// ErrColumnNotInTemplate 所选栏目不属于所选模板
	ErrColumnNotInTemplate = errors.New("所选栏目不属于该模板，请重新选择")
)

// ValidateTemplateColumn 校验「模板 + 栏目」归属（广告、友链等）：
//   - templateId、columnId 均必填（前端表单已标为必填，后端同样强制，避免绕过前端写入脏数据）；
//   - columnId 必须存在且其 template_id 等于 templateId。
func ValidateTemplateColumn(templateID, columnID uint) error {
	if templateID == 0 {
		return ErrTemplateRequired
	}
	if columnID == 0 {
		return ErrColumnRequired
	}

	var count int64
	if err := utils.DB.Model(&models.Column{}).
		Where("id = ? AND template_id = ?", columnID, templateID).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return ErrColumnNotInTemplate
	}
	return nil
}
