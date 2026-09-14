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
	return db.DB.Model(&models.ProjectBatch{}).Where("id = ?", id).Updates(clean).Error
}

func (s *BatchService) Delete(id uint64) error {
	var count int64
	db.DB.Model(&models.Application{}).Where("batch_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("该批次下存在申报记录，无法删除")
	}
	return db.DB.Delete(&models.ProjectBatch{}, id).Error
}

func (s *BatchService) Publish(id uint64) error {
	return db.DB.Model(&models.ProjectBatch{}).Where("id = ?", id).Update("status", models.BatchStatusOpen).Error
}

func (s *BatchService) Close(id uint64) error {
	return db.DB.Model(&models.ProjectBatch{}).Where("id = ?", id).Update("status", models.BatchStatusClosed).Error
}

func (s *BatchService) StartReview(id uint64) error {
	return db.DB.Model(&models.ProjectBatch{}).Where("id = ?", id).Update("status", models.BatchStatusReviewing).Error
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
