package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"

	"portal/internal/models"
	"portal/pkg/utils"
)

var contentSyncLocks = struct {
	sync.Mutex
	running map[uint]bool
}{running: make(map[uint]bool)}

type ContentSourceView struct {
	*models.ContentSource
	SecretConfigured bool `json:"secretConfigured"`
}

type ContentSourceService struct {
	wechat *WeChatContentProvider
}

func NewContentSourceService() *ContentSourceService {
	return &ContentSourceService{wechat: &WeChatContentProvider{}}
}

func (s *ContentSourceService) List(name string, status, page, pageSize int) ([]ContentSourceView, int64, error) {
	var sources []models.ContentSource
	var total int64
	query := utils.DB.Model(&models.ContentSource{})
	if strings.TrimSpace(name) != "" {
		query = query.Where("name LIKE ?", "%"+strings.TrimSpace(name)+"%")
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Limit(pageSize).Offset(offset).Find(&sources).Error; err != nil {
		return nil, 0, err
	}
	views := make([]ContentSourceView, 0, len(sources))
	for i := range sources {
		views = append(views, contentSourceView(&sources[i]))
	}
	return views, total, nil
}

func (s *ContentSourceService) Get(id uint) (*models.ContentSource, error) {
	var source models.ContentSource
	if err := utils.DB.First(&source, id).Error; err != nil {
		return nil, err
	}
	return &source, nil
}

func (s *ContentSourceService) Create(source *models.ContentSource) error {
	if err := normalizeAndValidateContentSource(source); err != nil {
		return err
	}
	if err := validateDefaultCategory(source.DefaultCategoryID); err != nil {
		return err
	}
	return utils.CreatePreservingZeroValues(utils.DB, source, "status")
}

func (s *ContentSourceService) Update(id uint, input *models.ContentSource) error {
	var old models.ContentSource
	if err := utils.DB.First(&old, id).Error; err != nil {
		return err
	}
	if strings.TrimSpace(input.AppSecret) == "" {
		input.AppSecret = old.AppSecret
	}
	if err := normalizeAndValidateContentSource(input); err != nil {
		return err
	}
	if err := validateDefaultCategory(input.DefaultCategoryID); err != nil {
		return err
	}
	updates := map[string]any{
		"name":                input.Name,
		"provider":            input.Provider,
		"account_type":        input.AccountType,
		"app_id":              input.AppID,
		"app_secret":          input.AppSecret,
		"secret_env":          "",
		"sync_mode":           input.SyncMode,
		"status":              input.Status,
		"default_category_id": input.DefaultCategoryID,
	}
	if err := utils.DB.Model(&old).Updates(updates).Error; err != nil {
		return err
	}
	s.wechat.InvalidateToken(id)
	return nil
}

func (s *ContentSourceService) Delete(id uint) error {
	var imports int64
	if err := utils.DB.Model(&models.ContentImport{}).Where("source_id = ?", id).Count(&imports).Error; err != nil {
		return err
	}
	if imports > 0 {
		return fmt.Errorf("该内容源已导入 %d 篇文章，请禁用而不要删除，以保留去重记录", imports)
	}
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("source_id = ?", id).Delete(&models.ContentSyncLog{}).Error; err != nil {
			return err
		}
		result := tx.Delete(&models.ContentSource{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		s.wechat.InvalidateToken(id)
		return nil
	})
}

func (s *ContentSourceService) TestConnection(id uint) error {
	source, err := s.Get(id)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.wechat.TestConnection(ctx, source)
}

func (s *ContentSourceService) StartSync(id, userID uint, operator string) (*models.ContentSyncLog, error) {
	source, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	if source.Status != 1 {
		return nil, fmt.Errorf("内容源已禁用")
	}
	contentSyncLocks.Lock()
	if contentSyncLocks.running[id] {
		contentSyncLocks.Unlock()
		return nil, fmt.Errorf("该内容源正在同步，请稍后查看同步记录")
	}
	contentSyncLocks.running[id] = true
	contentSyncLocks.Unlock()

	log := &models.ContentSyncLog{
		SourceID: id, TriggeredBy: userID, Operator: truncateRunes(operator, 100),
		Status: models.ContentSyncStatusQueued,
	}
	if err := utils.DB.Create(log).Error; err != nil {
		s.unlockSync(id)
		return nil, err
	}
	if err := utils.DB.Model(&models.ContentSource{}).Where("id = ?", id).Updates(map[string]any{
		"last_sync_status":  models.ContentSyncStatusQueued,
		"last_sync_message": "任务已提交",
	}).Error; err != nil {
		utils.Logger.Warnf("更新内容源同步状态失败: %s", err)
	}
	go s.runSync(id, log.ID)
	return log, nil
}

func (s *ContentSourceService) GetSyncLog(id uint) (*models.ContentSyncLog, error) {
	var log models.ContentSyncLog
	if err := utils.DB.First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func (s *ContentSourceService) ListSyncLogs(sourceID uint, page, pageSize int) ([]models.ContentSyncLog, int64, error) {
	var logs []models.ContentSyncLog
	var total int64
	query := utils.DB.Model(&models.ContentSyncLog{}).Where("source_id = ?", sourceID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Order("id DESC").Limit(pageSize).Offset(offset).Find(&logs).Error
	return logs, total, err
}

func (s *ContentSourceService) runSync(sourceID, logID uint) {
	defer s.unlockSync(sourceID)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	now := time.Now()
	_ = utils.DB.Model(&models.ContentSyncLog{}).Where("id = ?", logID).Updates(map[string]any{
		"status": models.ContentSyncStatusRunning, "started_at": &now,
	}).Error
	_ = utils.DB.Model(&models.ContentSource{}).Where("id = ?", sourceID).Updates(map[string]any{
		"last_sync_status":  models.ContentSyncStatusRunning,
		"last_sync_message": "正在从公众号读取文章",
	}).Error

	source, err := s.Get(sourceID)
	if err != nil {
		s.finishSync(logID, sourceID, 0, 0, 0, err)
		return
	}
	created, skipped := 0, 0
	scanned, err := s.wechat.FetchArticles(ctx, source, func(external ExternalContentArticle) error {
		if external.Deleted || strings.TrimSpace(external.Title) == "" {
			skipped++
			return nil
		}
		wasCreated, err := s.importArticle(source, external)
		if err != nil {
			return err
		}
		if wasCreated {
			created++
		} else {
			skipped++
		}
		return nil
	})
	s.finishSync(logID, sourceID, scanned, created, skipped, err)
}

func (s *ContentSourceService) importArticle(source *models.ContentSource, external ExternalContentArticle) (bool, error) {
	created := false
	err := utils.DB.Transaction(func(tx *gorm.DB) error {
		var existing models.ContentImport
		err := tx.Where("source_id = ? AND external_id = ? AND item_index = ?", source.ID, external.ExternalID, external.ItemIndex).
			First(&existing).Error
		if err == nil {
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		author := strings.TrimSpace(external.Author)
		if author == "" {
			author = source.Name
		}
		publishTime := models.LocalTime{Time: external.ExternalUpdatedAt}
		article := &models.Article{
			Title:       truncateRunes(strings.TrimSpace(external.Title), 200),
			Type:        models.ArticleTypeGraphic,
			Summary:     truncateRunes(strings.TrimSpace(external.Summary), 500),
			Content:     external.Content,
			Status:      models.ArticleStatusDraft,
			AuditStatus: 0,
			Cover:       truncateRunes(external.Cover, 500),
			Author:      truncateRunes(author, 100),
			AuthorCode:  strconv.FormatUint(uint64(source.CreatedBy), 10),
			Source:      truncateRunes(source.Name, 200),
			PublishTime: &publishTime,
			URL:         truncateRunes(external.URL, 500),
		}
		if err := tx.Create(article).Error; err != nil {
			return err
		}
		if source.DefaultCategoryID != nil && *source.DefaultCategoryID > 0 {
			join := &models.ArticleCategory{ArticleID: article.ID, CategoryID: *source.DefaultCategoryID}
			if err := tx.Create(join).Error; err != nil {
				return err
			}
		}
		externalTime := external.ExternalUpdatedAt
		mapping := &models.ContentImport{
			SourceID: source.ID, ExternalID: external.ExternalID, ItemIndex: external.ItemIndex,
			ArticleID: article.ID, ExternalUpdatedAt: &externalTime,
		}
		if err := tx.Create(mapping).Error; err != nil {
			return err
		}
		created = true
		return nil
	})
	return created, err
}

func (s *ContentSourceService) finishSync(logID, sourceID uint, scanned, created, skipped int, syncErr error) {
	finished := time.Now()
	status := models.ContentSyncStatusSuccess
	message := fmt.Sprintf("同步完成：扫描 %d 篇，新建 %d 篇，跳过 %d 篇", scanned, created, skipped)
	errorMessage := ""
	if syncErr != nil {
		status = models.ContentSyncStatusFailed
		errorMessage = truncateRunes(syncErr.Error(), 1000)
		message = "同步失败：" + errorMessage
		utils.Logger.Errorf("公众号内容同步失败 source_id=%d log_id=%d: %s", sourceID, logID, syncErr)
	}
	_ = utils.DB.Model(&models.ContentSyncLog{}).Where("id = ?", logID).Updates(map[string]any{
		"status": status, "finished_at": &finished, "scanned_count": scanned,
		"created_count": created, "skipped_count": skipped, "error_message": errorMessage,
	}).Error
	_ = utils.DB.Model(&models.ContentSource{}).Where("id = ?", sourceID).Updates(map[string]any{
		"last_sync_at": &finished, "last_sync_status": status,
		"last_sync_message": truncateRunes(message, 500),
	}).Error
}

func (s *ContentSourceService) unlockSync(sourceID uint) {
	contentSyncLocks.Lock()
	delete(contentSyncLocks.running, sourceID)
	contentSyncLocks.Unlock()
}

func normalizeAndValidateContentSource(source *models.ContentSource) error {
	source.Name = strings.TrimSpace(source.Name)
	source.Provider = strings.TrimSpace(source.Provider)
	source.AccountType = strings.TrimSpace(source.AccountType)
	source.AppID = strings.TrimSpace(source.AppID)
	source.AppSecret = strings.TrimSpace(source.AppSecret)
	source.SecretEnv = ""
	source.SyncMode = strings.TrimSpace(source.SyncMode)
	if source.Name == "" || source.AppID == "" || source.AppSecret == "" {
		return fmt.Errorf("名称、AppID 和 AppSecret 均不能为空")
	}
	if source.Provider != models.ContentSourceProviderWeChat {
		return fmt.Errorf("暂不支持该内容源类型")
	}
	if source.AccountType != models.ContentSourceAccountService && source.AccountType != models.ContentSourceAccountSubscription {
		return fmt.Errorf("公众号类型无效")
	}
	if source.SyncMode != models.ContentSourceSyncPublished && source.SyncMode != models.ContentSourceSyncMaterials {
		return fmt.Errorf("采集范围无效")
	}
	if len(source.AppSecret) > 255 {
		return fmt.Errorf("AppSecret 长度不能超过 255 个字符")
	}
	if source.Status != 0 && source.Status != 1 {
		return fmt.Errorf("状态无效")
	}
	if source.DefaultCategoryID != nil && *source.DefaultCategoryID == 0 {
		source.DefaultCategoryID = nil
	}
	return nil
}

func validateDefaultCategory(id *uint) error {
	if id == nil || *id == 0 {
		return nil
	}
	var count int64
	if err := utils.DB.Model(&models.Category{}).Where("id = ? AND status = 1", *id).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("默认分类不存在或已禁用")
	}
	return nil
}

func contentSourceView(source *models.ContentSource) ContentSourceView {
	return ContentSourceView{
		ContentSource:    source,
		SecretConfigured: strings.TrimSpace(source.AppSecret) != "",
	}
}

func truncateRunes(value string, max int) string {
	if max <= 0 || utf8.RuneCountInString(value) <= max {
		return value
	}
	runes := []rune(value)
	return string(runes[:max])
}
