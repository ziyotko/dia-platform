package models

import "time"

// LiveConfig for meeting live streaming
type LiveConfig struct {
	BaseModel
	MeetingID uint64     `gorm:"uniqueIndex" json:"meetingId"`
	StreamKey string     `gorm:"size:128" json:"streamKey"`
	PushURL   string     `gorm:"size:512" json:"pushUrl"`
	PlayURL   string     `gorm:"size:512" json:"playUrl"`
	Status    string     `gorm:"size:32;default:off" json:"status"` // off/live/ended
	StartedAt *time.Time `json:"startedAt"`
	EndedAt   *time.Time `json:"endedAt"`

	Meeting *Meeting `gorm:"foreignKey:MeetingID" json:"meeting,omitempty"`
}

func (LiveConfig) TableName() string {
	return "conference_live_configs"
}

// VodConfig for video-on-demand / replay
type VodConfig struct {
	BaseModel
	MeetingID    uint64     `gorm:"uniqueIndex" json:"meetingId"`
	VideoURL     string     `gorm:"size:512" json:"videoUrl"`
	AllowReplay  bool       `gorm:"default:false" json:"allowReplay"`
	ReplayExpiry *time.Time `json:"replayExpiry"`
	Status       string     `gorm:"size:32;default:draft" json:"status"`

	Meeting *Meeting `gorm:"foreignKey:MeetingID" json:"meeting,omitempty"`
}

func (VodConfig) TableName() string {
	return "conference_vod_configs"
}

// LiveMessage for live chat
type LiveMessage struct {
	BaseModel
	MeetingID uint64 `gorm:"index" json:"meetingId"`
	UserID    uint64 `gorm:"index" json:"userId"`
	Content   string `gorm:"type:text" json:"content"`

	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Meeting *Meeting `gorm:"foreignKey:MeetingID" json:"meeting,omitempty"`
}

func (LiveMessage) TableName() string {
	return "conference_live_messages"
}

// ViewingLog tracks watching duration
type ViewingLog struct {
	BaseModel
	MeetingID uint64     `gorm:"index" json:"meetingId"`
	UserID    uint64     `gorm:"index" json:"userId"`
	StartTime *time.Time `json:"startTime"`
	EndTime   *time.Time `json:"endTime"`
	Duration  int        `gorm:"default:0" json:"duration"` // seconds
	Type      string     `gorm:"size:32" json:"type"`       // live/vod

	Meeting *Meeting `gorm:"foreignKey:MeetingID" json:"meeting,omitempty"`
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (ViewingLog) TableName() string {
	return "conference_viewing_logs"
}
