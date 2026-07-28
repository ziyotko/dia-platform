package models

// ArticleCategory represents article category
type ArticleCategory struct {
	BaseModel
	Name string `gorm:"size:64;not null" json:"name"`
	Sort int    `gorm:"default:0" json:"sort"`
}

func (ArticleCategory) TableName() string {
	return "member_article_categories"
}
