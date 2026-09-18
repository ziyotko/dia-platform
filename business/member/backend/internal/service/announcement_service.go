package service

import (
	"errors"
	"member/internal/models"
	"member/pkg/db"
	"member/pkg/utils"
	"time"
)

type AnnouncementService struct{}

// GetPublishedAnnouncements returns published announcements (public)
func (s *AnnouncementService) GetPublishedAnnouncements(page, size int, keyword, aType string) ([]models.Announcement, int64, error) {
	var list []models.Announcement
	var total int64

	query := db.DB.Model(&models.Announcement{}).Where("published_at IS NOT NULL AND published_at <= ?", time.Now())
	if keyword != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if aType != "" {
		query = query.Where("type = ?", aType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("is_pinned DESC, published_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetAnnouncement returns an announcement by ID
func (s *AnnouncementService) GetAnnouncement(id uint64) (*models.Announcement, error) {
	var a models.Announcement
	if err := db.DB.First(&a, id).Error; err != nil {
		return nil, errors.New("公告不存在")
	}
	// Increment view count
	db.DB.Model(&a).UpdateColumn("view_count", a.ViewCount+1)
	return &a, nil
}

// CreateAnnouncement creates an announcement (admin)
func (s *AnnouncementService) CreateAnnouncement(req CreateAnnouncementRequest) (*models.Announcement, error) {
	a := models.Announcement{
		Title:     req.Title,
		Content:   utils.SanitizeRichText(req.Content),
		Type:      req.Type,
		IsPinned:  req.IsPinned,
		CreatedBy: req.CreatedBy,
	}
	if req.PublishNow {
		now := time.Now()
		a.PublishedAt = &models.LocalTime{Time: now}
	}
	if err := db.DB.Create(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

// UpdateAnnouncement updates an announcement (admin)
func (s *AnnouncementService) UpdateAnnouncement(id uint64, req UpdateAnnouncementRequest) error {
	updates := map[string]interface{}{
		"title":     req.Title,
		"content":   utils.SanitizeRichText(req.Content),
		"type":      req.Type,
		"is_pinned": req.IsPinned,
	}
	if req.PublishNow {
		now := time.Now()
		updates["published_at"] = &models.LocalTime{Time: now}
	}
	return db.DB.Model(&models.Announcement{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteAnnouncement deletes an announcement (admin)
func (s *AnnouncementService) DeleteAnnouncement(id uint64) error {
	return db.DB.Delete(&models.Announcement{}, id).Error
}

// ListAllAnnouncements lists all announcements (admin)
func (s *AnnouncementService) ListAllAnnouncements(page, size int) ([]models.Announcement, int64, error) {
	var list []models.Announcement
	var total int64

	query := db.DB.Model(&models.Announcement{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

type CreateAnnouncementRequest struct {
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Type       string `json:"type"`
	IsPinned   bool   `json:"is_pinned"`
	PublishNow bool   `json:"publish_now"`
	CreatedBy  string `json:"created_by"`
}

type UpdateAnnouncementRequest struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	Type       string `json:"type"`
	IsPinned   bool   `json:"is_pinned"`
	PublishNow bool   `json:"publish_now"`
}
