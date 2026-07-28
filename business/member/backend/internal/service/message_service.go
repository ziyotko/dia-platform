package service

import (
	"errors"
	"member/internal/models"
	"member/pkg/db"
	"time"
)

type MessageService struct{}

// CreateMessage creates a member message
func (s *MessageService) CreateMessage(memberID uint64, req CreateMessageRequest) (*models.MemberMessage, error) {
	msg := models.MemberMessage{
		MemberID: memberID,
		Title:    req.Title,
		Content:  req.Content,
		Status:   models.MessageStatusUnread,
	}
	if err := db.DB.Create(&msg).Error; err != nil {
		return nil, err
	}
	return &msg, nil
}

// GetMyMessages returns the member's messages
func (s *MessageService) GetMyMessages(memberID uint64, page, size int) ([]models.MemberMessage, int64, error) {
	var msgs []models.MemberMessage
	var total int64

	query := db.DB.Model(&models.MemberMessage{}).Where("member_id = ?", memberID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&msgs).Error; err != nil {
		return nil, 0, err
	}
	return msgs, total, nil
}

// GetMessage returns a message by ID
func (s *MessageService) GetMessage(id uint64) (*models.MemberMessage, error) {
	var msg models.MemberMessage
	if err := db.DB.Preload("Member").First(&msg, id).Error; err != nil {
		return nil, errors.New("消息不存在")
	}
	return &msg, nil
}

// ListAllMessages lists all messages (admin)
func (s *MessageService) ListAllMessages(page, size int, status string) ([]models.MemberMessage, int64, error) {
	var msgs []models.MemberMessage
	var total int64

	query := db.DB.Model(&models.MemberMessage{}).Preload("Member")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&msgs).Error; err != nil {
		return nil, 0, err
	}
	return msgs, total, nil
}

// ReplyMessage replies to a message (admin)
func (s *MessageService) ReplyMessage(id uint64, reply string) error {
	now := time.Now()
	return db.DB.Model(&models.MemberMessage{}).Where("id = ?", id).Updates(map[string]interface{}{
		"reply":      reply,
		"replied_at": &models.LocalTime{Time: now},
		"status":     models.MessageStatusReplied,
	}).Error
}

type CreateMessageRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}
