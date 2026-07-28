package models

// ArticleStatus constants
const (
	ArticleStatusDraft     = "draft"     // 草稿
	ArticleStatusPending   = "pending"   // 待审核
	ArticleStatusPublished = "published" // 已发布
	ArticleStatusRejected  = "rejected"  // 已拒绝
)

// Article represents a member-published article
type Article struct {
	BaseModel
	MemberID      uint64          `gorm:"index;not null" json:"member_id"`
	Member        Member          `gorm:"foreignKey:MemberID" json:"member,omitempty"`
	CategoryID    uint64          `gorm:"index" json:"category_id"`
	Category      ArticleCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Title         string          `gorm:"size:255;not null" json:"title"`
	Content       string          `gorm:"type:longtext" json:"content"`
	Summary       string          `gorm:"size:500" json:"summary"`
	CoverImage    string          `gorm:"size:255" json:"cover_image"`
	Status        string          `gorm:"size:20;default:draft" json:"status"`
	ViewCount     int             `gorm:"default:0" json:"view_count"`
	PublishedAt   *LocalTime      `gorm:"type:datetime" json:"published_at"`
	ReviewComment string          `gorm:"size:500" json:"review_comment"`
}

func (Article) TableName() string {
	return "member_articles"
}
