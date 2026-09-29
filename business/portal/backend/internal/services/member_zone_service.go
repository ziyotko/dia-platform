package services

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"portal/internal/models"
	"portal/pkg/utils"
)

// ============================ 会员栏目（会员栏目分类） ============================

type MemberColumnService struct{}

func (s *MemberColumnService) GetMemberColumns(name string, status int, page int, pageSize int) ([]models.MemberColumn, int64, error) {
	var list []models.MemberColumn
	var total int64
	query := utils.DB.Model(&models.MemberColumn{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("sort ASC, id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list).Error
	return list, total, err
}

func (s *MemberColumnService) CreateMemberColumn(column *models.MemberColumn) error {
	// Create 接口忽略请求体主键（防伪造内置 ID），与其它模块 Create 口径一致
	column.ID = 0
	if err := ensureNameCodeUnique(&models.MemberColumn{}, "会员栏目", column.Name, column.Code, 0, nil); err != nil {
		return err
	}
	return utils.DB.Create(column).Error
}

func (s *MemberColumnService) UpdateMemberColumn(id uint, column *models.MemberColumn) error {
	var old models.MemberColumn
	if err := utils.DB.First(&old, id).Error; err != nil {
		return err
	}
	if err := ensureNameCodeUnique(&models.MemberColumn{}, "会员栏目", column.Name, column.Code, id, nil); err != nil {
		return err
	}
	return utils.DB.Model(&old).Updates(map[string]any{
		"name":        column.Name,
		"code":        column.Code,
		"description": column.Description,
		"sort":        column.Sort,
		"status":      column.Status,
	}).Error
}

func (s *MemberColumnService) UpdateMemberColumnStatus(id uint, status int) error {
	return utils.DB.Model(&models.MemberColumn{}).Where("id = ?", id).Update("status", status).Error
}

// DeleteMemberColumn 删除会员栏目：栏目下仍有内容时拒绝，避免残留内容指向不存在的栏目。
func (s *MemberColumnService) DeleteMemberColumn(id uint) error {
	var column models.MemberColumn
	if err := utils.DB.First(&column, id).Error; err != nil {
		return err
	}
	var contentCount int64
	if err := utils.DB.Model(&models.MemberContent{}).Where("member_column_id = ?", id).Count(&contentCount).Error; err != nil {
		return err
	}
	if contentCount > 0 {
		return fmt.Errorf("该会员栏目下仍有 %d 条内容，请先删除或转移这些内容后再删除", contentCount)
	}
	return utils.DB.Delete(&column).Error
}

// ============================ 会员专属内容 ============================

type MemberContentService struct{}

func (s *MemberContentService) GetMemberContents(title string, columnID int, contentType int, status int, authorCodeScope string, page int, pageSize int) ([]models.MemberContent, int64, error) {
	var list []models.MemberContent
	var total int64
	query := utils.DB.Model(&models.MemberContent{})
	if title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}
	if columnID > 0 {
		query = query.Where("member_column_id = ?", columnID)
	}
	if contentType > 0 {
		query = query.Where("type = ?", contentType)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	// 非管理员只能看到自己发布的内容（内容作者仅限自有内容，与文章列表口径一致）
	if authorCodeScope != "" {
		query = query.Where("author_code = ?", authorCodeScope)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("is_top DESC, id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list).Error
	return list, total, err
}

func (s *MemberContentService) GetMemberContentByID(id uint) (*models.MemberContent, error) {
	var content models.MemberContent
	if err := utils.DB.First(&content, id).Error; err != nil {
		return nil, err
	}
	return &content, nil
}

func (s *MemberContentService) CreateMemberContent(content *models.MemberContent) error {
	content.ID = 0
	// 新建时必须投放到「启用中」的会员栏目
	if err := validateMemberContent(content, true); err != nil {
		return err
	}
	return utils.DB.Create(content).Error
}

func (s *MemberContentService) UpdateMemberContent(id uint, content *models.MemberContent) error {
	var old models.MemberContent
	if err := utils.DB.First(&old, id).Error; err != nil {
		return err
	}
	// 仅在「更换了会员栏目」时要求新栏目处于启用状态；原栏目被禁用后仍允许继续编辑内容
	if err := validateMemberContent(content, content.MemberColumnID != old.MemberColumnID); err != nil {
		return err
	}
	return utils.DB.Model(&old).Updates(map[string]any{
		"member_column_id": content.MemberColumnID,
		"title":            content.Title,
		"type":             content.Type,
		"summary":          content.Summary,
		"content":          content.Content,
		"cover":            content.Cover,
		"video_url":        content.VideoURL,
		"attachment_name":  content.AttachmentName,
		"attachment_url":   content.AttachmentURL,
		"source":           content.Source,
		"publish_time":     content.PublishTime,
		"status":           content.Status,
		"is_top":           content.IsTop,
	}).Error
}

func (s *MemberContentService) UpdateMemberContentStatus(id uint, status int) error {
	if status != models.MemberContentStatusDraft && status != models.MemberContentStatusPublished && status != models.MemberContentStatusOffline {
		return errors.New("状态值无效")
	}
	return utils.DB.Model(&models.MemberContent{}).Where("id = ?", id).Update("status", status).Error
}

func (s *MemberContentService) DeleteMemberContent(id uint) error {
	return utils.DB.Delete(&models.MemberContent{}, id).Error
}

// GetMemberColumnNames 按 ID 批量取会员栏目名称（列表展示用，避免 N+1 查询）。
func (s *MemberContentService) GetMemberColumnNames(ids []uint) (map[uint]string, error) {
	result := make(map[uint]string, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	var columns []models.MemberColumn
	if err := utils.DB.Where("id IN ?", ids).Find(&columns).Error; err != nil {
		return nil, err
	}
	for _, column := range columns {
		result[column.ID] = column.Name
	}
	return result, nil
}

// validateMemberContent 内容写入前的公共校验：标题/类型白名单/所属会员栏目有效性。
func validateMemberContent(content *models.MemberContent, requireEnabledColumn bool) error {
	if strings.TrimSpace(content.Title) == "" {
		return errors.New("内容标题不能为空")
	}
	if !models.IsValidMemberContentType(content.Type) {
		return errors.New("内容类型无效")
	}
	var column models.MemberColumn
	if err := utils.DB.First(&column, content.MemberColumnID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("所属会员栏目不存在")
		}
		return err
	}
	if requireEnabledColumn && column.Status != 1 {
		return fmt.Errorf("会员栏目「%s」已禁用，不能投放内容", column.Name)
	}
	return nil
}
