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

// MemberContent 会员专属内容，按业务类型使用不同字段组合（未使用的类型字段留空）：
//
//	新闻(1)：title + cover + content + attachment_name/url
//	数据(2)：title + data_year + unit_name + province + region + is_belt + is_axis + sub_field + main_business_income
//	视频(3)：title + cover + full_video_url + preview_video_url
//	报刊(4)：title + issue_no + publish_year_month + cover + summary + paper_file_name/url
//
// 共用字段：member_column_id（所属会员栏目）、type、source（来源）、publish_time（发布时间）、status、is_top。
type MemberContent struct {
	ID             uint       `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	MemberColumnID uint       `gorm:"not null;index" json:"memberColumnId"`
	Title          string     `gorm:"size:200;not null" json:"title"`
	Type           int        `gorm:"default:1;index" json:"type"` // 1新闻 2数据 3视频 4报刊
	Source         string     `gorm:"size:200" json:"source"`
	PublishTime    *LocalTime `json:"publishTime"`
	Status         int        `gorm:"default:0;index" json:"status"` // 0草稿 1已发布 2已下线
	IsTop          int        `gorm:"default:0" json:"isTop"`
	Author         string     `gorm:"size:100" json:"author"`
	AuthorCode     string     `gorm:"size:100;index" json:"authorCode"`
	ViewCount      int        `gorm:"default:0" json:"viewCount"`

	// —— 新闻(1) / 视频(3) / 报刊(4) 共用：封面图 ——
	Cover string `gorm:"size:500" json:"cover"`

	// —— 新闻(1) ——
	Content        string `gorm:"type:longtext" json:"content"`   // 文章内容（富文本 HTML）
	AttachmentName string `gorm:"size:255" json:"attachmentName"` // 文章附件名称
	AttachmentURL  string `gorm:"size:500" json:"attachmentUrl"`  // 文章附件地址

	// —— 数据(2) ——
	DataYear           string  `gorm:"size:20" json:"dataYear"`                                // 数据年份
	UnitName           string  `gorm:"size:200" json:"unitName"`                               // 单位名称
	Province           string  `gorm:"size:100" json:"province"`                               // 所属省份及直辖市
	Region             string  `gorm:"size:100" json:"region"`                                 // 所属地区
	IsBelt             int     `gorm:"default:0" json:"isBelt"`                                // 是否一带：0否 1是
	IsAxis             int     `gorm:"default:0" json:"isAxis"`                                // 是否一轴：0否 1是
	SubField           string  `gorm:"size:200" json:"subField"`                               // 细分领域
	MainBusinessIncome float64 `gorm:"type:decimal(18,2);default:0" json:"mainBusinessIncome"` // 主营业务收入（亿元）

	// —— 视频(3) ——
	FullVideoURL    string `gorm:"size:500" json:"fullVideoUrl"`    // 完整视频地址
	PreviewVideoURL string `gorm:"size:500" json:"previewVideoUrl"` // 预览视频地址

	// —— 报刊(4) ——
	IssueNo          string `gorm:"size:50" json:"issueNo"`          // 期号
	PublishYearMonth string `gorm:"size:20" json:"publishYearMonth"` // 出版年月（YYYY-MM）
	Summary          string `gorm:"size:500" json:"summary"`         // 摘要
	PaperFileName    string `gorm:"size:255" json:"paperFileName"`   // 报刊文件名称
	PaperFileURL     string `gorm:"size:500" json:"paperFileUrl"`    // 报刊文件地址
}

// MemberContent 类型取值（前端 member-zone.vue 的 typeName 与此一致）
const (
	MemberContentTypeNews  = 1 // 新闻
	MemberContentTypeData  = 2 // 数据
	MemberContentTypeVideo = 3 // 视频
	MemberContentTypePaper = 4 // 报刊
)

// IsValidMemberContentType 判断会员内容类型是否在白名单内
func IsValidMemberContentType(t int) bool {
	return t >= MemberContentTypeNews && t <= MemberContentTypePaper
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
