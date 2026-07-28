package models

// Announcement represents system announcements
type Announcement struct {
	BaseModel
	Title       string     `gorm:"size:255;not null" json:"title"`
	Content     string     `gorm:"type:text;not null" json:"content"`
	Type        string     `gorm:"size:20;default:notice" json:"type"` // notice/article/policy
	IsPinned    bool       `gorm:"default:false" json:"is_pinned"`
	ViewCount   int        `gorm:"default:0" json:"view_count"`
	PublishedAt *LocalTime `gorm:"type:datetime" json:"published_at"`
	CreatedBy   string     `gorm:"size:64" json:"created_by"`
}

func (Announcement) TableName() string {
	return "member_announcements"
}
