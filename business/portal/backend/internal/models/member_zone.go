package models

import (
	"encoding/json"
	"time"
)

// MemberColumn 会员栏目（会员栏目分类）。
// 会员专区下用于归类「会员专属内容」的栏目，仅管理员可维护；禁用后不再允许向其投放新内容。
type MemberColumn struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Code        string    `gorm:"size:100;not null" json:"code"`
	Description string    `gorm:"size:500" json:"description"`
	Sort        int       `gorm:"default:0" json:"sort"`
	Status      int       `gorm:"default:1;index" json:"status"` // 1启用 0禁用
}

func (m MemberColumn) MarshalJSON() ([]byte, error) {
	type Alias MemberColumn
	return json.Marshal(&struct {
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
		*Alias
	}{
		CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: m.UpdatedAt.Format("2006-01-02 15:04:05"),
		Alias:     (*Alias)(&m),
	})
}

// MemberContent 会员专属内容（新闻/数据/视频）。
// 每条内容归属一个会员栏目（MemberColumnID），由作者本人创建维护（管理员可代管全部）。
type MemberContent struct {
	ID             uint       `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	MemberColumnID uint       `gorm:"not null;index" json:"memberColumnId"`
	Title          string     `gorm:"size:200;not null" json:"title"`
	Type           int        `gorm:"default:1;index" json:"type"` // 1新闻 2数据 3视频
	Summary        string     `gorm:"size:500" json:"summary"`
	Content        string     `gorm:"type:longtext" json:"content"`
	Cover          string     `gorm:"size:500" json:"cover"`
	VideoURL       string     `gorm:"size:500" json:"videoUrl"`
	AttachmentName string     `gorm:"size:255" json:"attachmentName"`
	AttachmentURL  string     `gorm:"size:500" json:"attachmentUrl"`
	Author         string     `gorm:"size:100" json:"author"`
	AuthorCode     string     `gorm:"size:100;index" json:"authorCode"`
	Source         string     `gorm:"size:200" json:"source"`
	PublishTime    *LocalTime `json:"publishTime"`
	Status         int        `gorm:"default:0;index" json:"status"` // 0草稿 1已发布 2已下线
	IsTop          int        `gorm:"default:0" json:"isTop"`
	ViewCount      int        `gorm:"default:0" json:"viewCount"`
}

// MemberContent 类型取值（前端 member-zone.vue 的 typeName 与此一致）
const (
	MemberContentTypeNews  = 1 // 新闻
	MemberContentTypeData  = 2 // 数据
	MemberContentTypeVideo = 3 // 视频
)

// IsValidMemberContentType 判断会员内容类型是否在白名单内
func IsValidMemberContentType(t int) bool {
	return t >= MemberContentTypeNews && t <= MemberContentTypeVideo
}

// MemberContent 状态取值
const (
	MemberContentStatusDraft     = 0 // 草稿
	MemberContentStatusPublished = 1 // 已发布
	MemberContentStatusOffline   = 2 // 已下线
)

func (m MemberContent) MarshalJSON() ([]byte, error) {
	type Alias MemberContent
	return json.Marshal(&struct {
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
		*Alias
	}{
		CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: m.UpdatedAt.Format("2006-01-02 15:04:05"),
		Alias:     (*Alias)(&m),
	})
}
