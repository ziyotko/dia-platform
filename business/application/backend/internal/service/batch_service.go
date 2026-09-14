package service

import (
	"errors"
	"time"

	"application/internal/models"
	"application/pkg/db"
)

type BatchService struct{}

func (s *BatchService) Create(b *models.ProjectBatch) error {
	if b.Title == "" {
		return errors.New("请填写批次名称")
	}
	if err := checkCategory(b.CategoryID); err != nil {
		return err
	}
	if b.ApplyStart != nil && b.ApplyEnd != nil && !b.ApplyEnd.After(*b.ApplyStart) {
		return errors.New("申报截止时间必须晚于开始时间")
	}
	b.Status = models.BatchStatusDraft
	return db.DB.Create(b).Error
}

func (s *BatchService) Update(id uint64, updates map[string]interface{}) error {
	clean := pickUpdates(updates, "title", "category_id", "description", "requirements",
		"apply_start", "apply_end", "review_deadline")
	normalizeTimeFields(clean, "apply_start", "apply_end", "review_deadline")
	if len(clean) == 0 {
		return nil
	}
	var b models.ProjectBatch
	if err := db.DB.First(&b, id).Error; err != nil {
		return errors.New("批次不存在")
	}
	// A published batch has already frozen the round it defines: its category is
	// the one every application must match ("项目类别须与申报批次一致") and its
	// window is what already accepted the submissions. Only a draft may change
	// them, otherwise existing applications would silently become inconsistent.
	if b.Status != models.BatchStatusDraft {
		for _, field := range []string{"category_id", "apply_start", "apply_end"} {
			if _, ok := clean[field]; ok {
				return errors.New("批次已发布，不能修改项目类别和申报时间")
			}
		}
	}
	if raw, ok := clean["title"]; ok {
		title, _ := raw.(string)
		if title == "" {
			return errors.New("请填写批次名称")
		}
	}
	if raw, ok := clean["category_id"]; ok {
		if err := checkCategory(toUint64(raw)); err != nil {
			return err
		}
	}
	// Validate the merged (not the incoming) window so a batch can never be
	// stored with a deadline that is before its start date.
	start := mergeTime(b.ApplyStart, clean, "apply_start")
	end := mergeTime(b.ApplyEnd, clean, "apply_end")
	if start != nil && end != nil && !end.After(*start) {
		return errors.New("申报截止时间必须晚于开始时间")
	}
	return db.DB.Model(&b).Updates(clean).Error
}

// mergeTime returns the value a time column will have after applying updates.
// A key absent from updates keeps its current value; an explicit null clears it.
func mergeTime(current *time.Time, updates map[string]interface{}, key string) *time.Time {
	raw, ok := updates[key]
	if !ok {
		return current
	}
	if t, ok := raw.(time.Time); ok {
		return &t
	}
	return nil
}

func (s *BatchService) Delete(id uint64) error {
	var count int64
	db.DB.Model(&models.Application{}).Where("batch_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("该批次下存在申报记录，无法删除")
	}
	var annCount int64
	db.DB.Model(&models.Announcement{}).Where("batch_id = ?", id).Count(&annCount)
	if annCount > 0 {
		return errors.New("该批次下存在结果公示，无法删除")
	}
	return db.DB.Delete(&models.ProjectBatch{}, id).Error
}

// Publish opens a draft batch for applications. A batch can only be published
// once and must have a category plus a complete application window, otherwise it
// would silently stay open for ever.
func (s *BatchService) Publish(id uint64) error {
	var b models.ProjectBatch
	if err := db.DB.First(&b, id).Error; err != nil {
		return errors.New("批次不存在")
	}
	if b.Status != models.BatchStatusDraft {
		return errors.New("仅草稿状态的批次可发布")
	}
	if b.Title == "" {
		return errors.New("请先填写批次名称")
	}
	if err := checkCategory(b.CategoryID); err != nil {
		return err
	}
	if b.ApplyStart == nil || b.ApplyEnd == nil {
		return errors.New("请先设置申报开始和截止时间")
	}
	if !b.ApplyEnd.After(*b.ApplyStart) {
		return errors.New("申报截止时间必须晚于开始时间")
	}
	return db.DB.Model(&b).Update("status", models.BatchStatusOpen).Error
}

// Close ends a batch that is either accepting applications or under review.
func (s *BatchService) Close(id uint64) error {
	var b models.ProjectBatch
	if err := db.DB.First(&b, id).Error; err != nil {
		return errors.New("批次不存在")
	}
	if b.Status != models.BatchStatusOpen && b.Status != models.BatchStatusReviewing {
		return errors.New("仅申报中或评审中的批次可结束")
	}
	return db.DB.Model(&b).Update("status", models.BatchStatusClosed).Error
}

// StartReview moves an open batch into the review stage.
func (s *BatchService) StartReview(id uint64) error {
	var b models.ProjectBatch
	if err := db.DB.First(&b, id).Error; err != nil {
		return errors.New("批次不存在")
	}
	if b.Status != models.BatchStatusOpen {
		return errors.New("仅申报中的批次可进入评审阶段")
	}
	return db.DB.Model(&b).Update("status", models.BatchStatusReviewing).Error
}

func (s *BatchService) GetByID(id uint64) (*models.ProjectBatch, error) {
	var b models.ProjectBatch
	if err := db.DB.Preload("Category").First(&b, id).Error; err != nil {
		return nil, errors.New("批次不存在")
	}
	return &b, nil
}

func (s *BatchService) List(page, size int, keyword, status string) ([]models.ProjectBatch, int64, error) {
	var list []models.ProjectBatch
	var total int64
	query := db.DB.Model(&models.ProjectBatch{})
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	err := query.Preload("Category").Order("created_at DESC").
		Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

// ListOpen returns batches currently open for application (frontend). The
// result honours the same time window as IsOpen so the list and the submit
// guard can never disagree.
func (s *BatchService) ListOpen(page, size int, keyword string) ([]models.ProjectBatch, int64, error) {
	var list []models.ProjectBatch
	var total int64
	now := time.Now()
	query := db.DB.Model(&models.ProjectBatch{}).Where("status = ?", models.BatchStatusOpen).
		Where("apply_start IS NULL OR apply_start <= ?", now).
		Where("apply_end IS NULL OR apply_end >= ?", now)
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Preload("Category").Order("created_at DESC").
		Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

// IsOpen checks whether the batch is still accepting applications
func (s *BatchService) IsOpen(b *models.ProjectBatch) bool {
	if b.Status != models.BatchStatusOpen {
		return false
	}
	now := time.Now()
	if b.ApplyStart != nil && now.Before(*b.ApplyStart) {
		return false
	}
	if b.ApplyEnd != nil && now.After(*b.ApplyEnd) {
		return false
	}
	return true
}

// IsReviewOpen checks whether the batch still accepts review work.
// A nil review deadline means "no deadline".
func (s *BatchService) IsReviewOpen(b *models.ProjectBatch) bool {
	if b.ReviewDeadline == nil {
		return true
	}
	return !time.Now().After(*b.ReviewDeadline)
}

// AutoCloseExpired moves open batches past their deadline into reviewing
func (s *BatchService) AutoCloseExpired() {
	db.DB.Model(&models.ProjectBatch{}).
		Where("status = ? AND apply_end IS NOT NULL AND apply_end < ?", models.BatchStatusOpen, time.Now()).
		Update("status", models.BatchStatusReviewing)
}
