package services

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"portal/config"
	"portal/internal/models"
	"portal/pkg/utils"
)

// ============================ 会员栏目（会员栏目分类） ============================

type MemberColumnService struct{}

// GetMemberColumnByKey 按「稳定标识」解析会员栏目，供对外接口使用：
//   - 纯数字：按 ID 匹配（兼容既有按 columnId 调用的方式）；
//   - 其它：按名称精确匹配（名称写入前已做唯一校验，故名称可作跨环境的稳定业务键）。
//
// 设计初衷：各环境的自增 ID 不一致，外部系统（如 CAMIE）按名称调用即可免去「ID 映射配置」。
func (s *MemberColumnService) GetMemberColumnByKey(key string) (*models.MemberColumn, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, errors.New("会员栏目标识不能为空")
	}
	var column models.MemberColumn
	var err error
	if id, parseErr := strconv.ParseUint(key, 10, 32); parseErr == nil {
		err = utils.DB.First(&column, uint(id)).Error
	} else {
		err = utils.DB.Where("name = ?", key).First(&column).Error
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("会员栏目不存在")
		}
		return nil, err
	}
	return &column, nil
}

// GetAllMemberColumns 返回全部会员栏目（含禁用），供对外接口发现「有哪些栏目」。
func (s *MemberColumnService) GetAllMemberColumns() ([]models.MemberColumn, error) {
	var list []models.MemberColumn
	err := utils.DB.Order("sort ASC, id DESC").Find(&list).Error
	return list, err
}

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
	// Status 字段带 `default:1` 标签：GORM 会把「零值」当作未设置而改用库默认值（1），
	// 且插入后还会把库里的默认值回填进结构体 →「新建时选禁用」会被静默存成启用。
	// 因此先记下本次要写的状态，插入后与库中不一致时显式补写一次。
	status := column.Status
	if err := utils.DB.Create(column).Error; err != nil {
		return err
	}
	if status == column.Status {
		return nil
	}
	if err := utils.DB.Model(&models.MemberColumn{}).Where("id = ?", column.ID).Update("status", status).Error; err != nil {
		return err
	}
	column.Status = status
	return nil
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

// MemberContentQuery 管理端「会员内容列表」的筛选条件。
// 约定：ColumnID/Type <= 0、Status/ColumnStatus < 0、字符串为空 均表示「该项不过滤」。
type MemberContentQuery struct {
	Title        string
	ColumnID     int
	Type         int
	Status       int    // 内容状态：0 草稿 / 1 已发布 / 2 已下线
	ColumnStatus int    // 所属会员栏目的启用状态：1 启用 / 0 禁用
	PublishStart string // 发布时间起（YYYY-MM-DD，含当天）
	PublishEnd   string // 发布时间止（YYYY-MM-DD，含当天）
	// 非管理员只查本人内容（按 author_code）；空 = 不限制
	AuthorCodeScope string
	Page            int
	PageSize        int
}

func (s *MemberContentService) GetMemberContents(q MemberContentQuery) ([]models.MemberContent, int64, error) {
	var list []models.MemberContent
	var total int64
	query := utils.DB.Model(&models.MemberContent{})
	if q.Title != "" {
		query = query.Where("title LIKE ?", "%"+q.Title+"%")
	}
	if q.ColumnID > 0 {
		query = query.Where("member_column_id = ?", q.ColumnID)
	}
	if q.Type > 0 {
		query = query.Where("type = ?", q.Type)
	}
	if q.Status >= 0 {
		query = query.Where("status = ?", q.Status)
	}
	// 按「所属会员栏目的启用状态」筛选：用子查询而不是 JOIN，
	// 这样 Count 语义与 SELECT 列保持不变（JOIN 会同时影响两者）。
	if q.ColumnStatus >= 0 {
		query = query.Where("member_column_id IN (?)",
			utils.DB.Model(&models.MemberColumn{}).Select("id").Where("status = ?", q.ColumnStatus))
	}
	// 发布时间区间：半开区间（含起始当天、含结束当天），只传一端也生效。
	// 未填发布时间（NULL）的内容会被区间条件排除，符合「按发布时间筛选」的预期。
	if start, ok := parseDateParam(q.PublishStart); ok {
		query = query.Where("publish_time >= ?", start)
	}
	if end, ok := parseDateParam(q.PublishEnd); ok {
		query = query.Where("publish_time < ?", end.AddDate(0, 0, 1))
	}
	// 非管理员只能看到自己发布的内容（内容作者仅限自有内容，与文章列表口径一致）
	if q.AuthorCodeScope != "" {
		query = query.Where("author_code = ?", q.AuthorCodeScope)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("is_top DESC, id DESC").Limit(q.PageSize).Offset((q.Page - 1) * q.PageSize).Find(&list).Error
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

// MemberFileReference 引用某私有文件的会员内容（用于「按 URL 反查」的访问判定）。
// 只取判定所需的最小字段，避免把正文等大字段读出来。
type MemberFileReference struct {
	ID         uint   `gorm:"column:id"`
	Status     int    `gorm:"column:status"`
	AuthorCode string `gorm:"column:author_code"`
}

// FindMemberFileReferences 反查引用了指定私有文件的会员内容。
//
// 文件名在库中有两种出现形式：
//   - 封面图/文章附件/完整视频/预览视频/报刊文件 → 整条 URL 存在对应列里（用等值匹配）；
//   - 正文内联图片 → 存在 `content` 的 HTML 里（用 LIKE 匹配，无法走索引，
//     但 member_content 量级很小且签发频率低，可接受；若日后变慢可加短 TTL 缓存）。
func (s *MemberContentService) FindMemberFileReferences(fileName string) ([]MemberFileReference, error) {
	canonical := config.AppConfig.Server.ApiPrefix + "/member-files/" + fileName
	inlineLike := "%/member-files/" + fileName + "%"
	var refs []MemberFileReference
	err := utils.DB.Model(&models.MemberContent{}).
		Select("id", "status", "author_code").
		Where("cover = ? OR attachment_url = ? OR full_video_url = ? OR preview_video_url = ? OR paper_file_url = ? OR content LIKE ?",
			canonical, canonical, canonical, canonical, canonical, inlineLike).
		Find(&refs).Error
	return refs, err
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
