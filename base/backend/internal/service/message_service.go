package service

import (
	"time"

	"base/internal/models"
	"base/pkg/db"
	"base/pkg/notifier"
)

type MessageService struct{}

func (s MessageService) Create(m *models.Message) error {
	if m.Status == 3 && m.SendAt == nil {
		now := time.Now()
		m.SendAt = &now
	}
	return db.DB.Create(m).Error
}

// MessageListQuery 消息列表查询条件。
// Box=inbox 收件箱（别人发给我的 + 广播）；Box=sent 发件箱（我发出的，按 sender_id）。
type MessageListQuery struct {
	TenantID uint64
	UserID   uint64
	Box      string
	Status   int // -1 全部
	IsRead   int // -1 全部，0 未读，1 已读
	Page     int
	Size     int
}

func (s MessageService) GetByID(id uint64, tenantID uint64) (*models.Message, error) {
	var m models.Message
	db := db.DB
	if tenantID > 0 {
		db = db.Where("tenant_id = ?", tenantID)
	}
	err := db.First(&m, id).Error
	return &m, err
}

// Delete 删除消息：租户用户只能删除本租户数据，平台超管（tenantID=0）可删除任意消息。
func (s MessageService) Delete(id uint64, tenantID uint64) error {
	query := db.DB.Where("id = ?", id)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	return query.Delete(&models.Message{}).Error
}

func (s MessageService) List(q MessageListQuery) ([]models.Message, int64, error) {
	var list []models.Message
	var total int64
	query := db.DB.Model(&models.Message{})
	if q.TenantID > 0 {
		query = query.Where("tenant_id = ? OR tenant_id = 0", q.TenantID)
	}
	if q.Box == "sent" {
		// 发件箱：我发出的消息（草稿与已发送均在列）
		query = query.Where("sender_id = ?", q.UserID)
	} else {
		// 收件箱：别人发给我或广播给我，且已发送（不含草稿）
		query = query.Where("receiver_id = ? OR receiver_id = 0", q.UserID).Where("status = ?", 3)
	}
	if q.Status >= 0 {
		query = query.Where("status = ?", q.Status)
	}
	if q.IsRead >= 0 {
		query = query.Where("is_read = ?", q.IsRead == 1)
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((q.Page - 1) * q.Size).Limit(q.Size).Find(&list).Error
	return list, total, err
}

// GetUnreadCount 当前用户未读数（收件箱中未读的已发送消息）。
func (s MessageService) GetUnreadCount(receiverID uint64, tenantID uint64) (int64, error) {
	var count int64
	query := db.DB.Model(&models.Message{}).
		Where("(receiver_id = ? OR receiver_id = 0) AND status = ? AND is_read = ?", receiverID, 3, false)
	if tenantID > 0 {
		query = query.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}
	err := query.Count(&count).Error
	return count, err
}

func (s MessageService) MarkRead(id uint64, receiverID uint64, tenantID uint64) error {
	updates := map[string]interface{}{
		"is_read": true,
		"read_at": time.Now(),
	}
	query := db.DB.Model(&models.Message{}).Where("id = ? AND (receiver_id = ? OR receiver_id = 0)", id, receiverID)
	if tenantID > 0 {
		query = query.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}
	return query.Updates(updates).Error
}

// MarkAllRead 将当前用户的全部未读消息标为已读，返回影响行数。
func (s MessageService) MarkAllRead(receiverID uint64, tenantID uint64) (int64, error) {
	query := db.DB.Model(&models.Message{}).
		Where("(receiver_id = ? OR receiver_id = 0) AND is_read = ?", receiverID, false)
	if tenantID > 0 {
		query = query.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}
	res := query.Updates(map[string]interface{}{
		"is_read": true,
		"read_at": time.Now(),
	})
	return res.RowsAffected, res.Error
}

func (s MessageService) SendToUsers(senderID uint64, senderName string, tenantID uint64, userIDs []uint64, title, content, msgType, priority string) error {
	now := time.Now()
	var msgs []models.Message
	for _, uid := range userIDs {
		msgs = append(msgs, models.Message{
			TenantID:     tenantID,
			SenderID:     senderID,
			SenderName:   senderName,
			ReceiverID:   uid,
			ReceiverType: "user",
			Title:        title,
			Content:      content,
			Type:         msgType,
			Priority:     priority,
			Status:       3,
			SendAt:       &now,
		})
	}
	return db.DB.CreateInBatches(msgs, 100).Error
}

func (s MessageService) Broadcast(senderID uint64, senderName string, tenantID uint64, title, content, msgType, priority string) error {
	now := time.Now()
	return db.DB.Create(&models.Message{
		TenantID:     tenantID,
		SenderID:     senderID,
		SenderName:   senderName,
		ReceiverID:   0,
		ReceiverType: "all",
		Title:        title,
		Content:      content,
		Type:         msgType,
		Priority:     priority,
		Status:       3,
		SendAt:       &now,
	}).Error
}

// SendExternal 根据渠道发送站外消息（邮件/短信/企微等）
func (s MessageService) SendExternal(channel, to, subject, content string) error {
	if channel == "" || channel == "in-app" {
		return nil
	}
	return notifier.Send(channel, notifier.Message{
		To:      to,
		Subject: subject,
		Body:    content,
	})
}
