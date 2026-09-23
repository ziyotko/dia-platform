package service

import (
	"errors"
	"member/internal/models"
	"member/pkg/db"
	"strings"
	"time"
)

type MessageService struct{}

// CreateMessage creates a member message
func (s *MessageService) CreateMessage(memberID uint64, req CreateMessageRequest) (*models.MemberMessage, error) {
	title := strings.TrimSpace(req.Title)
	content := strings.TrimSpace(req.Content)
	// 长度前置校验：title 列是 varchar(255)，超长会直接返回 MySQL 1406 原文给前端
	if title == "" {
		return nil, errors.New("请填写留言标题")
	}
	if len([]rune(title)) > 200 {
		return nil, errors.New("留言标题不能超过 200 字")
	}
	if content == "" {
		return nil, errors.New("请填写留言内容")
	}
	if len([]rune(content)) > 5000 {
		return nil, errors.New("留言内容不能超过 5000 字")
	}

	msg := models.MemberMessage{
		MemberID: memberID,
		Title:    title,
		Content:  content,
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

// GetMessage returns a message by ID (owner only).
// 会员查看留言时自动标记为“已读”（已回复的消息保持 replied）。
func (s *MessageService) GetMessage(id, memberID uint64) (*models.MemberMessage, error) {
	var msg models.MemberMessage
	if err := db.DB.Preload("Member").Where("id = ? AND member_id = ?", id, memberID).First(&msg).Error; err != nil {
		return nil, errors.New("消息不存在")
	}
	if msg.Status == models.MessageStatusUnread {
		if err := db.DB.Model(&models.MemberMessage{}).Where("id = ?", msg.ID).
			Update("status", models.MessageStatusRead).Error; err == nil {
			msg.Status = models.MessageStatusRead
		}
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
	var msg models.MemberMessage
	if err := db.DB.First(&msg, id).Error; err != nil {
		return errors.New("留言不存在")
	}
	now := time.Now()
	return db.DB.Model(&models.MemberMessage{}).Where("id = ?", id).Updates(map[string]interface{}{
		"reply":      reply,
		"replied_at": &models.LocalTime{Time: now},
		"status":     models.MessageStatusReplied,
	}).Error
}

// DeleteMessage deletes a message (admin)
func (s *MessageService) DeleteMessage(id uint64) error {
	return db.DB.Delete(&models.MemberMessage{}, id).Error
}

type CreateMessageRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}
