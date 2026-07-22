package service

import (
	"time"

	"base/internal/models"
	"base/pkg/db"
	"base/pkg/notifier"
)

type MessageService struct{}

func (s MessageService) Create(m *models.Message) error {
	if m.Status == 3 && m.SendAt.IsZero() {
		m.SendAt = time.Now()
	}
	return db.DB.Create(m).Error
}

func (s MessageService) Update(m *models.Message, tenantID uint64) error {
	db := db.DB.Model(m)
	if tenantID > 0 {
		db = db.Where("tenant_id = ?", tenantID)
	}
	return db.Updates(map[string]interface{}{
		"title":         m.Title,
		"content":       m.Content,
		"type":          m.Type,
		"priority":      m.Priority,
		"receiver_id":   m.ReceiverID,
		"receiver_type": m.ReceiverType,
		"status":        m.Status,
	}).Error
}

func (s MessageService) Delete(id uint64, tenantID uint64) error {
	db := db.DB
	if tenantID > 0 {
		db = db.Where("tenant_id = ?", tenantID)
	}
	return db.Delete(&models.Message{BaseModel: models.BaseModel{ID: id}}).Error
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

func (s MessageService) List(tenantID, receiverID uint64, status int, page, size int) ([]models.Message, int64, error) {
	var list []models.Message
	var total int64
	query := db.DB.Model(&models.Message{})
	if tenantID > 0 {
		query = query.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}
	if receiverID > 0 {
		query = query.Where("receiver_id = ? OR receiver_id = 0", receiverID)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (s MessageService) GetUnreadCount(receiverID uint64) (int64, error) {
	var count int64
	err := db.DB.Model(&models.Message{}).Where("(receiver_id = ? OR receiver_id = 0) AND status = ?", receiverID, 0).Count(&count).Error
	return count, err
}

func (s MessageService) MarkRead(id uint64, receiverID uint64) error {
	updates := map[string]interface{}{
		"status":  1,
		"read_at": time.Now(),
	}
	return db.DB.Model(&models.Message{}).Where("id = ? AND (receiver_id = ? OR receiver_id = 0)", id, receiverID).Updates(updates).Error
}

func (s MessageService) Send(m *models.Message) error {
	m.Status = 3
	m.SendAt = time.Now()
	return db.DB.Create(m).Error
}

func (s MessageService) SendToUsers(senderID uint64, senderName string, userIDs []uint64, title, content, msgType, priority string) error {
	now := time.Now()
	var msgs []models.Message
	for _, uid := range userIDs {
		msgs = append(msgs, models.Message{
			SenderID:     senderID,
			SenderName:   senderName,
			ReceiverID:   uid,
			ReceiverType: "user",
			Title:        title,
			Content:      content,
			Type:         msgType,
			Priority:     priority,
			Status:       3,
			SendAt:       now,
		})
	}
	return db.DB.CreateInBatches(msgs, 100).Error
}

func (s MessageService) Broadcast(senderID uint64, senderName string, tenantID uint64, title, content, msgType, priority string) error {
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
		SendAt:       time.Now(),
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
