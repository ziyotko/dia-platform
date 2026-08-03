package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Meeting represents a conference/meeting
type Meeting struct {
	BaseModel
	Title       string     `gorm:"size:256" json:"title"`
	Type        string     `gorm:"size:32;default:offline" json:"type"` // online/offline/hybrid
	StartTime   *time.Time `json:"startTime"`
	EndTime     *time.Time `json:"endTime"`
	Location    string     `gorm:"size:512" json:"location"`
	Capacity    int        `gorm:"default:0" json:"capacity"` // 0 = unlimited
	Fee         float64    `gorm:"type:decimal(10,2);default:0" json:"fee"`
	CoverImage  string     `gorm:"size:512" json:"coverImage"`
	Status      string     `gorm:"size:32;default:draft" json:"status"`
	Description string     `gorm:"type:text" json:"description"`

	// Feature switches
	IsPaid        bool `gorm:"default:false" json:"isPaid"`
	NeedApproval  bool `gorm:"default:false" json:"needApproval"`
	AllowWaitlist bool `gorm:"default:false" json:"allowWaitlist"`
	EnableVote    bool `gorm:"default:false" json:"enableVote"`
	EnableSurvey  bool `gorm:"default:false" json:"enableSurvey"`
	EnableCredit  bool `gorm:"default:false" json:"enableCredit"`
	EnableLive    bool `gorm:"default:false" json:"enableLive"`

	// Time constraints
	RegStartTime   *time.Time `json:"regStartTime"`
	RegEndTime     *time.Time `json:"regEndTime"`
	SignStartTime  *time.Time `json:"signStartTime"`
	SignEndTime    *time.Time `json:"signEndTime"`
	CancelDeadline *time.Time `json:"cancelDeadline"`

	// Access control
	AccessLevels   JSONArray `gorm:"type:json" json:"accessLevels"`
	AccessBranches JSONArray `gorm:"type:json" json:"accessBranches"`

	// Relations
	Agendas []Agenda `gorm:"foreignKey:MeetingID" json:"agendas,omitempty"`
	Guests  []Guest  `gorm:"foreignKey:MeetingID" json:"guests,omitempty"`
}

func (Meeting) TableName() string {
	return "conference_meetings"
}

// Agenda item for a meeting
type Agenda struct {
	BaseModel
	MeetingID uint64     `gorm:"index" json:"meetingId"`
	Title     string     `gorm:"size:256" json:"title"`
	Speaker   string     `gorm:"size:128" json:"speaker"`
	StartTime *time.Time `json:"startTime"`
	EndTime   *time.Time `json:"endTime"`
	Sort      int        `gorm:"default:0" json:"sort"`
}

func (Agenda) TableName() string {
	return "conference_agendas"
}

// Guest/Speaker for a meeting
type Guest struct {
	BaseModel
	MeetingID uint64 `gorm:"index" json:"meetingId"`
	Name      string `gorm:"size:128" json:"name"`
	Title     string `gorm:"size:256" json:"title"`
	Avatar    string `gorm:"size:512" json:"avatar"`
	Bio       string `gorm:"size:1024" json:"bio"`
	Sort      int    `gorm:"default:0" json:"sort"`
}

func (Guest) TableName() string {
	return "conference_guests"
}

// JSONArray for JSON column support
type JSONArray []string

func (j JSONArray) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}
