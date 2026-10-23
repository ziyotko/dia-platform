package models

import "time"

// MemberStatisticsDataset 是会员专区「数据中心」的一期年度统计数据。
// 导入时先保存为草稿，避免尚未核对的 Excel 自动对会员发布。
type MemberStatisticsDataset struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	MemberColumnID uint      `gorm:"not null;uniqueIndex:idx_member_statistics_column_year" json:"memberColumnId"`
	StatisticsYear int       `gorm:"not null;uniqueIndex:idx_member_statistics_column_year" json:"statisticsYear"`
	ComparisonYear int       `gorm:"not null" json:"comparisonYear"`
	Title          string    `gorm:"size:200;not null" json:"title"`
	SourceFileName string    `gorm:"size:255;not null" json:"sourceFileName"`
	SourceSHA256   string    `gorm:"size:64;not null" json:"sourceSha256"`
	ImportedAt     time.Time `gorm:"not null" json:"importedAt"`
	Status         int       `gorm:"default:0;index" json:"status"` // 0草稿 1已发布 2已下线
}

const (
	MemberStatisticsStatusDraft     = 0
	MemberStatisticsStatusPublished = 1
	MemberStatisticsStatusOffline   = 2
)

// MemberStatisticsRow 保留 Excel「总表」中的一条原始样本。
// 企业名称可能重复，因此唯一键使用数据集 + 原始 Excel 行号，不按企业名去重。
// 金额原单位为万元；比例存 0~1 的原始数值，不提前换算或四舍五入到亿元。
type MemberStatisticsRow struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	DatasetID       uint      `gorm:"not null;uniqueIndex:idx_member_statistics_dataset_row;index" json:"datasetId"`
	SourceRow       int       `gorm:"not null;uniqueIndex:idx_member_statistics_dataset_row" json:"sourceRow"`
	EnterpriseName  string    `gorm:"size:255;not null;index" json:"enterpriseName"`
	SubField        string    `gorm:"size:100;not null;index" json:"subField"`
	Region          string    `gorm:"size:50;not null;index" json:"region"`
	IsBelt          int       `gorm:"default:0" json:"isBelt"`
	IsAxis          int       `gorm:"default:0" json:"isAxis"`
	Province        string    `gorm:"size:50;not null;index" json:"province"`
	RevenueYearWan  string    `gorm:"type:decimal(28,10);not null" json:"revenueYearWan"`
	RevenuePriorWan string    `gorm:"type:decimal(28,10);not null" json:"revenuePriorWan"`
	YoYSource       *string   `gorm:"type:decimal(20,15)" json:"yoySource"`
	RDYearWan       *string   `gorm:"type:decimal(28,10)" json:"rdYearWan"`
	RDPriorWan      *string   `gorm:"type:decimal(28,10)" json:"rdPriorWan"`
	RDShareYear     *string   `gorm:"type:decimal(20,15)" json:"rdShareYear"`
	RDSharePrior    *string   `gorm:"type:decimal(20,15)" json:"rdSharePrior"`
	Source          string    `gorm:"size:200" json:"source"`
}
