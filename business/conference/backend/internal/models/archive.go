package models

import "time"

// Archive represents an archived meeting
type Archive struct {
	BaseModel
	MeetingID    uint64     `gorm:"uniqueIndex" json:"meetingId"`
	MeetingTitle string     `gorm:"size:256" json:"meetingTitle"`
	MeetingType  string     `gorm:"size:32" json:"meetingType"`
	MeetingYear  int        `gorm:"index" json:"meetingYear"`
	ArchivedAt   *time.Time `json:"archivedAt"`
	OperatorID   uint64     `json:"operatorId"`

	Items []ArchiveItem `gorm:"foreignKey:ArchiveID" json:"items,omitempty"`
}

func (Archive) TableName() string {
	return "conference_archives"
}

// ArchiveItem is a single archived artifact
type ArchiveItem struct {
	BaseModel
	ArchiveID   uint64 `gorm:"index" json:"archiveId"`
	ItemType    string `gorm:"size:32" json:"itemType"` // registration/signin/finance/vote/survey/material/minutes/live
	FileName    string `gorm:"size:256" json:"fileName"`
	FilePath    string `gorm:"size:512" json:"filePath"`
	ContentJSON string `gorm:"type:longtext" json:"contentJson"`
	FileSize    int64  `gorm:"default:0" json:"fileSize"`
}

func (ArchiveItem) TableName() string {
	return "conference_archive_items"
}
