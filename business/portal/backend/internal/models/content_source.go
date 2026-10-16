package models

import "time"

const (
	ContentSourceProviderWeChat = "wechat_official"

	ContentSourceAccountService      = "service"
	ContentSourceAccountSubscription = "subscription"

	ContentSourceSyncPublished = "published"
	ContentSourceSyncMaterials = "materials"

	ContentSyncStatusQueued  = "queued"
	ContentSyncStatusRunning = "running"
	ContentSyncStatusSuccess = "success"
	ContentSyncStatusFailed  = "failed"
)

// ContentSource 只保存密钥的环境变量名。AppSecret 是兼容旧表结构的废弃列，不再用于同步。
type ContentSource struct {
	ID                uint       `gorm:"primarykey" json:"id"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	Name              string     `gorm:"size:100;not null" json:"name"`
	Provider          string     `gorm:"size:32;not null;uniqueIndex:uk_content_source_provider_app,priority:1" json:"provider"`
	AccountType       string     `gorm:"size:32;not null" json:"accountType"`
	AppID             string     `gorm:"size:64;not null;uniqueIndex:uk_content_source_provider_app,priority:2" json:"appId"`
	SecretEnv         string     `gorm:"size:128;not null" json:"secretEnv"`
	AppSecret         string     `gorm:"size:255" json:"-"` // 旧数据列；保存账号时清空
	SyncMode          string     `gorm:"size:32;not null" json:"syncMode"`
	Status            int        `gorm:"default:1;index" json:"status"`
	DefaultCategoryID *uint      `gorm:"index" json:"defaultCategoryId"`
	CreatedBy         uint       `gorm:"index" json:"createdBy"`
	LastSyncAt        *time.Time `json:"lastSyncAt"`
	LastSyncStatus    string     `gorm:"size:20" json:"lastSyncStatus"`
	LastSyncMessage   string     `gorm:"size:500" json:"lastSyncMessage"`
}

// ContentImport 建立外部文章与本地文章的一对一映射，用于幂等去重。
// ItemIndex 用于公众号一条多图文消息中的第 N 篇文章。
type ContentImport struct {
	ID                uint       `gorm:"primarykey" json:"id"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	SourceID          uint       `gorm:"not null;uniqueIndex:uk_content_import_external,priority:1;index" json:"sourceId"`
	ExternalID        string     `gorm:"size:191;not null;uniqueIndex:uk_content_import_external,priority:2" json:"externalId"`
	ItemIndex         int        `gorm:"not null;uniqueIndex:uk_content_import_external,priority:3" json:"itemIndex"`
	ArticleID         uint       `gorm:"not null;uniqueIndex" json:"articleId"`
	ExternalUpdatedAt *time.Time `json:"externalUpdatedAt"`
}

// ContentSyncLog 记录每次异步采集任务的状态与结果。
type ContentSyncLog struct {
	ID           uint       `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	SourceID     uint       `gorm:"not null;index" json:"sourceId"`
	TriggeredBy  uint       `gorm:"index" json:"triggeredBy"`
	Operator     string     `gorm:"size:100" json:"operator"`
	Status       string     `gorm:"size:20;not null;index" json:"status"`
	StartedAt    *time.Time `json:"startedAt"`
	FinishedAt   *time.Time `json:"finishedAt"`
	ScannedCount int        `json:"scannedCount"`
	CreatedCount int        `json:"createdCount"`
	SkippedCount int        `json:"skippedCount"`
	ErrorMessage string     `gorm:"size:1000" json:"errorMessage"`
}
