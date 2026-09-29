package services

import (
	"errors"
	"fmt"
	"regexp"
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
	if err := ensureValueUnique(&models.MemberColumn{}, "name", column.Name, "会员栏目", "名称", 0, nil); err != nil {
		return err
	}
	return utils.DB.Create(column).Error
}

func (s *MemberColumnService) UpdateMemberColumn(id uint, column *models.MemberColumn) error {
	var old models.MemberColumn
	if err := utils.DB.First(&old, id).Error; err != nil {
		return err
	}
	if err := ensureValueUnique(&models.MemberColumn{}, "name", column.Name, "会员栏目", "名称", id, nil); err != nil {
		return err
	}
	return utils.DB.Model(&old).Updates(map[string]any{
		"name":        column.Name,
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

// GetPublishedMemberContentsByColumn 对外接口用：分页取指定会员栏目下「已发布」的内容（置顶优先）。
// 栏目不存在时直接报错，便于调用方区分「栏目无内容」与「栏目不存在」。
func (s *MemberContentService) GetPublishedMemberContentsByColumn(columnID uint, page int, pageSize int) ([]models.MemberContent, int64, error) {
	var column models.MemberColumn
	if err := utils.DB.First(&column, columnID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, errors.New("会员栏目不存在")
		}
		return nil, 0, err
	}
	query := utils.DB.Model(&models.MemberContent{}).
		Where("member_column_id = ? AND status = ?", columnID, models.MemberContentStatusPublished)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.MemberContent
	err := query.Order("is_top DESC, id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list).Error
	return list, total, err
}

// GetPublishedMemberContentByID 对外接口用：按 ID 取「已发布」内容的完整信息。
// 草稿/已下线内容一律按「不存在」处理（对外只暴露已发布内容）。
func (s *MemberContentService) GetPublishedMemberContentByID(id uint) (*models.MemberContent, error) {
	var content models.MemberContent
	err := utils.DB.Where("id = ? AND status = ?", id, models.MemberContentStatusPublished).First(&content).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("内容不存在或未发布")
		}
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
		"member_column_id":     content.MemberColumnID,
		"title":                content.Title,
		"type":                 content.Type,
		"source":               content.Source,
		"publish_time":         content.PublishTime,
		"status":               content.Status,
		"is_top":               content.IsTop,
		"cover":                content.Cover,
		"content":              content.Content,
		"attachment_name":      content.AttachmentName,
		"attachment_url":       content.AttachmentURL,
		"data_year":            content.DataYear,
		"unit_name":            content.UnitName,
		"province":             content.Province,
		"region":               content.Region,
		"is_belt":              content.IsBelt,
		"is_axis":              content.IsAxis,
		"sub_field":            content.SubField,
		"main_business_income": content.MainBusinessIncome,
		"full_video_url":       content.FullVideoURL,
		"preview_video_url":    content.PreviewVideoURL,
		"issue_no":             content.IssueNo,
		"publish_year_month":   content.PublishYearMonth,
		"summary":              content.Summary,
		"paper_file_name":      content.PaperFileName,
		"paper_file_url":       content.PaperFileURL,
	}).Error
}

func (s *MemberContentService) UpdateMemberContentStatus(id uint, status int) error {
	if status != models.MemberContentStatusDraft && status != models.MemberContentStatusPublished && status != models.MemberContentStatusOffline {
		return errors.New("状态值无效")
	}
	// 切到「已发布」时同样要过该类型的业务必填项校验，避免绕过编辑保存发布空壳内容
	var content models.MemberContent
	if err := utils.DB.First(&content, id).Error; err != nil {
		return err
	}
	content.Status = status
	if err := validateMemberContentByType(&content); err != nil {
		return err
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
	return validateMemberContentByType(content)
}

// htmlTagRE 用于把富文本正文转成纯文本（空编辑器产出 <p><br></p>，直接判空会误判为有内容）
var htmlTagRE = regexp.MustCompile(`<[^>]*>`)

// memberContentPlainText 去掉 HTML 标签与 &nbsp; 后返回去掉首尾空白的纯文本
func memberContentPlainText(html string) string {
	plain := htmlTagRE.ReplaceAllString(html, "")
	plain = strings.ReplaceAll(plain, "&nbsp;", " ")
	return strings.TrimSpace(plain)
}

// validateMemberContentByType 按业务类型校验「已发布」内容的关键字段。
// 草稿（0）与已下线（2）不做限制：允许先把内容建好、逐步补齐后再发布。
func validateMemberContentByType(content *models.MemberContent) error {
	if content.Status != models.MemberContentStatusPublished {
		return nil
	}
	switch content.Type {
	case models.MemberContentTypeNews:
		// 富文本空编辑器会产出 <p><br></p>，需转纯文本后再判空
		if memberContentPlainText(content.Content) == "" && strings.TrimSpace(content.AttachmentURL) == "" {
			return errors.New("发布新闻前请填写文章内容或上传文章附件")
		}
	case models.MemberContentTypeData:
		if strings.TrimSpace(content.DataYear) == "" {
			return errors.New("发布数据前请填写数据年份")
		}
		if strings.TrimSpace(content.UnitName) == "" {
			return errors.New("发布数据前请填写单位名称")
		}
	case models.MemberContentTypeVideo:
		if strings.TrimSpace(content.FullVideoURL) == "" {
			return errors.New("发布视频前请上传完整视频")
		}
	case models.MemberContentTypePaper:
		if strings.TrimSpace(content.IssueNo) == "" {
			return errors.New("发布报刊前请填写期号")
		}
		if strings.TrimSpace(content.PublishYearMonth) == "" {
			return errors.New("发布报刊前请填写出版年月")
		}
		if strings.TrimSpace(content.PaperFileURL) == "" {
			return errors.New("发布报刊前请上传报刊文件")
		}
	}
	return nil
}
