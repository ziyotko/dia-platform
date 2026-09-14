package services

import (
	"errors"

	"server/models"
	"server/utils"
)

var (
	// ErrPageRequired 未选择所属页面
	ErrPageRequired = errors.New("请选择所属页面")
	// ErrColumnRequired 未选择所属栏目
	ErrColumnRequired = errors.New("请选择所属栏目")
	// ErrColumnNotInPage 所选栏目不属于所选页面
	ErrColumnNotInPage = errors.New("所选栏目不属于该页面，请重新选择")
)

// ValidatePageColumn 校验「页面 + 栏目」归属（广告、友链等）：
//   - pageId、columnId 均必填（前端表单已标为必填，后端同样强制，避免绕过前端写入脏数据）；
//   - columnId 必须存在且其 page_id 等于 pageId。
func ValidatePageColumn(pageID, columnID uint) error {
	if pageID == 0 {
		return ErrPageRequired
	}
	if columnID == 0 {
		return ErrColumnRequired
	}

	var count int64
	if err := utils.DB.Model(&models.Column{}).
		Where("id = ? AND page_id = ?", columnID, pageID).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return ErrColumnNotInPage
	}
	return nil
}
