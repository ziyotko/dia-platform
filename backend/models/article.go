package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Article struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"createTime"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Title        string         `gorm:"size:200;not null" json:"title"`
	CategoryID   uint           `gorm:"not null;index" json:"categoryId"`
	Category     Category       `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Summary      string         `gorm:"size:500" json:"summary"`
	Content      string         `gorm:"type:longtext" json:"content"`
	Status       int            `gorm:"default:0" json:"status"`      // 0草稿 1已发布 2已下线
	AuditStatus  int            `gorm:"default:0" json:"auditStatus"` // 0待审核 1审核中 2已审核
	IsTop        int            `gorm:"default:0" json:"isTop"`
	IsBold       int            `gorm:"default:0" json:"isBold"`
	DefaultColor string         `gorm:"size:20" json:"defaultColor"`
	Cover        string         `gorm:"size:500" json:"cover"`
	Author       string         `gorm:"size:100" json:"author"`
	AuthorCode   string         `gorm:"size:100" json:"authorCode"`
	Source       string         `gorm:"size:200" json:"source"`
	ColumnCount  int            `gorm:"column:column_count;default:0" json:"columnCount"`
	Tags         []Tag          `gorm:"many2many:article_tags;" json:"tags,omitempty"`
	Columns      []Column       `gorm:"many2many:article_columns;" json:"columns,omitempty"`
}

func (a Article) MarshalJSON() ([]byte, error) {
	type Alias Article
	return json.Marshal(&struct {
		CreatedAt string `json:"createTime"`
		UpdatedAt string `json:"updatedAt"`
		*Alias
	}{
		CreatedAt: a.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: a.UpdatedAt.Format("2006-01-02 15:04:05"),
		Alias:     (*Alias)(&a),
	})
}
