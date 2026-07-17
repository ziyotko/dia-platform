package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type LocalTime struct {
	time.Time
}

func (t *LocalTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" {
		t.Time = time.Time{}
		return nil
	}
	parsed, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

func (t LocalTime) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + t.Format("2006-01-02 15:04:05") + `"`), nil
}

func (t LocalTime) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}
	return t.Time, nil
}

func (t *LocalTime) Scan(value interface{}) error {
	if value == nil {
		t.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		t.Time = v
		return nil
	case string:
		parsed, err := time.Parse("2006-01-02 15:04:05", v)
		if err != nil {
			return err
		}
		t.Time = parsed
		return nil
	default:
		return fmt.Errorf("cannot scan %T into LocalTime", value)
	}
}

func (LocalTime) GORMDataType() string {
	return "datetime"
}

type Article struct {
	ID           uint                `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time           `gorm:"index:idx_article_status_audit_created;index:idx_article_author_status_created" json:"createTime"`
	UpdatedAt    time.Time           `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt      `gorm:"index" json:"deletedAt"`
	Title        string              `gorm:"size:200;not null" json:"title"`
	Type         int                 `gorm:"default:1;index" json:"type"` // 1图文 2视频 3数据
	Summary      string              `gorm:"size:500" json:"summary"`
	Content      string              `gorm:"type:longtext" json:"content"`
	Status       int                 `gorm:"default:0;index:idx_article_status_audit_created;index:idx_article_author_status_created" json:"status"` // 0草稿 1已发布 2已下线
	AuditStatus  int                 `gorm:"default:0;index:idx_article_status_audit_created" json:"auditStatus"`                                    // 0待审核 1审核中 2已审核
	IsTop        int                 `gorm:"default:0;index" json:"isTop"`
	IsBold       int                 `gorm:"default:0" json:"isBold"`
	DefaultColor string              `gorm:"size:20" json:"defaultColor"`
	Cover        string              `gorm:"size:500" json:"cover"`
	Author       string              `gorm:"size:100" json:"author"`
	AuthorCode   string              `gorm:"size:100;index:idx_article_author_status_created" json:"authorCode"`
	Source       string              `gorm:"size:200" json:"source"`
	PublishTime  *LocalTime          `json:"publishTime"`
	URL          string              `gorm:"size:500" json:"url"`
	ColumnCount  int                 `gorm:"column:column_count;default:0" json:"columnCount"`
	Categories   []Category          `gorm:"many2many:article_category;" json:"categories,omitempty"`
	Tags         []Tag               `gorm:"many2many:article_tag;" json:"tags,omitempty"`
	Columns      []Column            `gorm:"many2many:article_column;" json:"columns,omitempty"`
	Attachments  []ArticleAttachment `gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE;" json:"attachments,omitempty"`
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
