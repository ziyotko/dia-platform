package models

// Expert represents a review expert in the expert database (专家库)
type Expert struct {
	BaseModel
	AdminID      uint64 `gorm:"index" json:"adminId"`
	Name         string `gorm:"size:64" json:"name"`
	Username     string `gorm:"size:64;uniqueIndex" json:"username"`
	Specialty    string `gorm:"size:256" json:"specialty"`
	Title        string `gorm:"size:64" json:"title"`
	Organization string `gorm:"size:256" json:"organization"`
	Phone        string `gorm:"size:32" json:"phone"`
	Email        string `gorm:"size:128" json:"email"`
	Bio          string `gorm:"type:text" json:"bio"`
	Status       int    `gorm:"default:1" json:"status"`

	ReviewCount int64 `gorm:"-" json:"reviewCount"`
	ScoredCount int64 `gorm:"-" json:"scoredCount"`
}

func (Expert) TableName() string {
	return "application_experts"
}
