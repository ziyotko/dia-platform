package service

import (
	"errors"
	"time"

	"application/internal/models"
	"application/pkg/db"
)

type NotificationService struct{}

// Send sends a notification to all applicants (userID = 0) or to a single one.
// Broadcast rows are inserted in a single statement so a partial broadcast can
// never be delivered.
func (s *NotificationService) Send(title, content, ntype string, userID uint64) error {
	if title == "" || content == "" {
		return errors.New("请填写通知标题和内容")
	}
	if userID > 0 {
		var user models.User
		if err := db.DB.First(&user, userID).Error; err != nil {
			return errors.New("申报人不存在")
		}
		return db.DB.Create(&models.Notification{UserID: userID, Title: title, Content: content, Type: ntype}).Error
	}
	var users []models.User
	if err := db.DB.Select("id").Find(&users).Error; err != nil {
		return err
	}
	if len(users) == 0 {
		return nil
	}
	rows := make([]models.Notification, 0, len(users))
	for _, u := range users {
		rows = append(rows, models.Notification{UserID: u.ID, Title: title, Content: content, Type: ntype})
	}
	return db.DB.Create(&rows).Error
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
