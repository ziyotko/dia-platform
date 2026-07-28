package models

// MessageStatus constants
const (
	MessageStatusUnread  = "unread"  // 未读
	MessageStatusRead    = "read"    // 已读
	MessageStatusReplied = "replied" // 已回复
)

// MemberMessage represents a message/feedback from member
type MemberMessage struct {
	BaseModel
	MemberID  uint64     `gorm:"index;not null" json:"member_id"`
	Member    Member     `gorm:"foreignKey:MemberID" json:"member,omitempty"`
	Title     string     `gorm:"size:255;not null" json:"title"`
	Content   string     `gorm:"type:text;not null" json:"content"`
	Reply     string     `gorm:"type:text" json:"reply"`
	RepliedAt *LocalTime `gorm:"type:datetime" json:"replied_at"`
	Status    string     `gorm:"size:20;default:unread" json:"status"`
}

func (MemberMessage) TableName() string {
	return "member_messages"
}
