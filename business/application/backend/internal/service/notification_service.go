package service

import (
	"errors"
	"time"

	"application/internal/models"
	"application/pkg/db"
)

type NotificationService struct{}

// Send sends a notification to all users or a specific user
func (s *NotificationService) Send(title, content, ntype string, userID uint64) error {
	if title == "" || content == "" {
		return errors.New("请填写通知标题和内容")
	}
	if userID > 0 {
		return db.DB.Create(&models.Notification{UserID: userID, Title: title, Content: content, Type: ntype}).Error
	}
	// broadcast to all users
	var users []models.User
	if err := db.DB.Select("id").Find(&users).Error; err != nil {
		return err
	}
	for _, u := range users {
		db.DB.Create(&models.Notification{UserID: u.ID, Title: title, Content: content, Type: ntype})
	}
	return nil
}

func (s *NotificationService) ListUser(userID uint64, page, size int) ([]models.Notification, int64, error) {
	var list []models.Notification
	var total int64
	query := db.DB.Model(&models.Notification{}).Where("user_id = ?", userID)
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (s *NotificationService) MarkRead(id, userID uint64) error {
	now := time.Now()
	return db.DB.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("read_at", now).Error
}

func (s *NotificationService) UnreadCount(userID uint64) int64 {
	var count int64
	db.DB.Model(&models.Notification{}).Where("user_id = ? AND read_at IS NULL", userID).Count(&count)
	return count
}
