package models

import "time"

// Vote definition
type Vote struct {
	BaseModel
	MeetingID   uint64     `gorm:"index" json:"meetingId"`
	Title       string     `gorm:"size:256" json:"title"`
	Type        string     `gorm:"size:32" json:"type"` // single/multiple/equal/differential
	IsAnonymous bool       `gorm:"default:false" json:"isAnonymous"`
	StartTime   *time.Time `json:"startTime"`
	EndTime     *time.Time `json:"endTime"`
	Status      string     `gorm:"size:32;default:draft" json:"status"` // draft/open/closed
	MinSelect   int        `gorm:"default:1" json:"minSelect"`
	MaxSelect   int        `gorm:"default:1" json:"maxSelect"`
	TargetScope string     `gorm:"size:512" json:"targetScope"` // JSON: scope config

	// Relations
	Meeting *Meeting     `gorm:"foreignKey:MeetingID" json:"meeting,omitempty"`
	Options []VoteOption `gorm:"foreignKey:VoteID" json:"options,omitempty"`
}

func (Vote) TableName() string {
	return "conference_votes"
}

// VoteOption represents a candidate or choice
type VoteOption struct {
	BaseModel
	VoteID        uint64 `gorm:"index" json:"voteId"`
	Label         string `gorm:"size:256" json:"label"`
	CandidateName string `gorm:"size:128" json:"candidateName"`
	Sort          int    `gorm:"default:0" json:"sort"`
	VoteCount     int    `gorm:"default:0" json:"voteCount"`
}

func (VoteOption) TableName() string {
	return "conference_vote_options"
}

// VoteRecord stores individual votes
type VoteRecord struct {
	BaseModel
	VoteID    uint64     `gorm:"index" json:"voteId"`
	UserID    uint64     `gorm:"index" json:"userId"`
	OptionID  uint64     `gorm:"default:0" json:"optionId"`  // single vote
	OptionIDs string     `gorm:"type:text" json:"optionIds"` // JSON array for multi-vote
	VotedAt   *time.Time `json:"votedAt"`

	// Relations
	Vote *Vote `gorm:"foreignKey:VoteID" json:"vote,omitempty"`
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (VoteRecord) TableName() string {
	return "conference_vote_records"
}
