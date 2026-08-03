package service

import (
	"encoding/json"
	"time"

	"conference/internal/models"
	"conference/pkg/db"
)

type NotificationService struct{}

// Send creates a notification for target users
func (s *NotificationService) Send(senderID uint64, title, content, targetType string, targetIDs []uint64) error {
	idsJSON, _ := json.Marshal(targetIDs)
	now := time.Now()
	notif := models.Notification{
		SenderID:   senderID,
		Title:      title,
		Content:    content,
		TargetType: targetType,
		TargetIDs:  string(idsJSON),
		SentAt:     &now,
	}
	return db.DB.Create(&notif).Error
}

// GetUserNotifications returns notifications for a user
func (s *NotificationService) GetUserNotifications(userID uint64, page, size int) ([]map[string]interface{}, int64, error) {
	var allNotifs []models.Notification
	db.DB.Order("created_at DESC").Find(&allNotifs)

	type result struct {
		Notification models.Notification
		ReadAt       *time.Time
	}

	var results []result
	for _, n := range allNotifs {
		// Check if user is in target
		if s.isUserInTarget(userID, n.TargetType, n.TargetIDs) {
			var read models.NotificationRead
			err := db.DB.Where("notification_id = ? AND user_id = ?", n.ID, userID).First(&read).Error
			var readAt *time.Time
			if err == nil {
				readAt = read.ReadAt
			}
			results = append(results, result{Notification: n, ReadAt: readAt})
		}
	}

	total := int64(len(results))
	start := (page - 1) * size
	if start > int(total-1) {
		start = int(total)
	}
	end := start + size
	if end > int(total) {
		end = int(total)
	}

	var output []map[string]interface{}
	for _, r := range results[start:end] {
		output = append(output, map[string]interface{}{
			"id":         r.Notification.ID,
			"title":      r.Notification.Title,
			"content":    r.Notification.Content,
			"sent_at":    r.Notification.SentAt,
			"created_at": r.Notification.CreatedAt,
			"is_read":    r.ReadAt != nil,
			"read_at":    r.ReadAt,
		})
	}
	return output, total, nil
}

// MarkRead marks a notification as read for a user
func (s *NotificationService) MarkRead(notificationID, userID uint64) error {
	var count int64
	db.DB.Model(&models.NotificationRead{}).Where("notification_id = ? AND user_id = ?",
		notificationID, userID).Count(&count)
	if count > 0 {
		return nil
	}
	now := time.Now()
	read := models.NotificationRead{
		NotificationID: notificationID,
		UserID:         userID,
		ReadAt:         &now,
	}
	return db.DB.Create(&read).Error
}

// GetUnreadCount returns unread count for a user
func (s *NotificationService) GetUnreadCount(userID uint64) (int64, error) {
	var allNotifs []models.Notification
	db.DB.Find(&allNotifs)

	var totalRelevant int64
	for _, n := range allNotifs {
		if s.isUserInTarget(userID, n.TargetType, n.TargetIDs) {
			totalRelevant++
		}
	}

	var readCount int64
	db.DB.Model(&models.NotificationRead{}).Where("user_id = ?", userID).Count(&readCount)

	unread := totalRelevant - readCount
	if unread < 0 {
		unread = 0
	}
	return unread, nil
}

// ListSent returns sent notifications (admin view)
func (s *NotificationService) ListSent(page, size int) ([]models.Notification, int64, error) {
	var list []models.Notification
	var total int64
	db.DB.Model(&models.Notification{}).Count(&total)
	err := db.DB.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

// GetReadReceipt returns read receipts for a notification
func (s *NotificationService) GetReadReceipt(notificationID uint64) (map[string]interface{}, error) {
	var reads []models.NotificationRead
	db.DB.Preload("Notification").Where("notification_id = ?", notificationID).Find(&reads)

	var sentCount int64
	db.DB.Model(&models.Notification{}).Where("id = ?", notificationID).Count(&sentCount)

	return map[string]interface{}{
		"sent_count": sentCount,
		"read_count": len(reads),
		"reads":      reads,
	}, nil
}

func (s *NotificationService) isUserInTarget(userID uint64, targetType, targetIDsJSON string) bool {
	if targetType == models.NotifyTargetAll {
		return true
	}

	var targetIDs []uint64
	json.Unmarshal([]byte(targetIDsJSON), &targetIDs)

	if targetType == models.NotifyTargetSpecific {
		for _, id := range targetIDs {
			if id == userID {
				return true
			}
		}
		return false
	}

	// For registered/branch/level, always include for now (simplified)
	return true
}

// --- Auto-notification triggers ---

// NotifyRegistrationResult notifies user of registration approval/rejection
func (s *NotificationService) NotifyRegistrationResult(userID uint64, meetingTitle, status string) {
	title := "报名审核结果"
	content := "您在会议【" + meetingTitle + "】的报名已"
	if status == models.RegStatusApproved {
		content += "通过审核"
	} else {
		content += "被驳回"
	}
	s.Send(0, title, content, models.NotifyTargetSpecific, []uint64{userID})
}

// NotifyWaitlistPromotion notifies user they've been promoted from waitlist
func (s *NotificationService) NotifyWaitlistPromotion(userID uint64, meetingTitle string) {
	s.Send(0, "候补转正通知", "您在会议【"+meetingTitle+"】已从候补转为正式报名", models.NotifyTargetSpecific, []uint64{userID})
}

// NotifyRefundResult notifies user of refund result
func (s *NotificationService) NotifyRefundResult(userID uint64, meetingTitle string, approved bool) {
	title := "退费结果通知"
	content := "您在会议【" + meetingTitle + "】的退费申请已"
	if approved {
		content += "通过"
	} else {
		content += "被拒绝"
	}
	s.Send(0, title, content, models.NotifyTargetSpecific, []uint64{userID})
}

// NotifyLiveReminder sends live start reminder to registered users
func (s *NotificationService) NotifyLiveReminder(meetingID uint64, meetingTitle string) {
	var regs []models.Registration
	db.DB.Where("meeting_id = ? AND status = ?", meetingID, models.RegStatusApproved).Find(&regs)
	var userIDs []uint64
	for _, r := range regs {
		userIDs = append(userIDs, r.UserID)
	}
	if len(userIDs) > 0 {
		s.Send(0, "直播开始提醒", "会议【"+meetingTitle+"】直播已开始，请及时观看", models.NotifyTargetSpecific, userIDs)
	}
}
