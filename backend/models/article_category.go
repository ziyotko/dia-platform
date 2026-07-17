package models

type ArticleCategory struct {
	ArticleID  uint     `gorm:"primaryKey" json:"articleId"`
	CategoryID uint     `gorm:"primaryKey;index" json:"categoryId"`
	Category   Category `gorm:"foreignKey:CategoryID" json:"category"`
}
