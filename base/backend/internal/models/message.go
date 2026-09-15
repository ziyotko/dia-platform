package models

import "time"

type Message struct {
	BaseModel
	TenantID     uint64 `gorm:"index;comment:租户ID，0表示平台级" json:"tenantId"`
	SenderID     uint64 `gorm:"index;comment:发送者ID，0表示系统" json:"senderId"`
	SenderName   string `gorm:"size:64;comment:发送者名称" json:"senderName"`
	ReceiverID   uint64 `gorm:"index;comment:接收者ID，0表示广播" json:"receiverId"`
	ReceiverType string `gorm:"size:32;comment:接收类型 user/role/all" json:"receiverType"`
	Title        string `gorm:"size:256;comment:标题" json:"title"`
	Content      string `gorm:"type:text;comment:内容" json:"content"`
	Type         string `gorm:"size:32;comment:类型 notice/system/private" json:"type"`
	Priority     string `gorm:"size:16;comment:优先级 low/normal/high/urgent" json:"priority"`
	Status       int    `gorm:"default:3;comment:状态 2草稿 3已发送" json:"status"`
	IsRead       bool   `gorm:"default:false;comment:是否已读" json:"isRead"`
	// ReadAt / SendAt 为指针类型：未读/未发送时写入 NULL，
	// 避免 GORM 写入 0000-00-00 被 MySQL 严格模式（NO_ZERO_DATE）拒绝
	ReadAt *time.Time `gorm:"comment:读取时间" json:"readAt"`
	SendAt *time.Time `gorm:"comment:发送时间" json:"sendAt"`
}

func (Message) TableName() string {
	return "base_message"
}

type MessageTemplate struct {
	BaseModel
	TenantID    uint64 `gorm:"index;comment:租户ID" json:"tenantId"`
	Code        string `gorm:"size:64;comment:模板编码" json:"code"`
	Name        string `gorm:"size:128;comment:模板名称" json:"name"`
	Channel     string `gorm:"size:32;comment:渠道 in-app/sms/email/wechat" json:"channel"`
	Subject     string `gorm:"size:256;comment:主题" json:"subject"`
	Content     string `gorm:"type:text;comment:内容" json:"content"`
	Variables   string `gorm:"type:text;comment:变量说明JSON" json:"variables"`
	Status      int    `gorm:"default:1;comment:状态" json:"status"`
	Description string `gorm:"size:512;comment:描述" json:"description"`
}

func (MessageTemplate) TableName() string {
	return "base_message_template"
}
