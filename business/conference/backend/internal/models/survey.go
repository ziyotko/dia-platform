package models

import "time"

// Survey definition
type Survey struct {
	BaseModel
	MeetingID   uint64     `gorm:"index" json:"meetingId"`
	Title       string     `gorm:"size:256" json:"title"`
	Description string     `gorm:"type:text" json:"description"`
	StartTime   *time.Time `json:"startTime"`
	EndTime     *time.Time `json:"endTime"`
	Status      string     `gorm:"size:32;default:draft" json:"status"` // draft/open/closed

	Meeting   *Meeting         `gorm:"foreignKey:MeetingID" json:"meeting,omitempty"`
	Questions []SurveyQuestion `gorm:"foreignKey:SurveyID" json:"questions,omitempty"`
}

func (Survey) TableName() string {
	return "conference_surveys"
}

// SurveyQuestion represents a question in a survey
type SurveyQuestion struct {
	BaseModel
	SurveyID uint64 `gorm:"index" json:"surveyId"`
	Title    string `gorm:"size:512" json:"title"`
	Type     string `gorm:"size:32" json:"type"` // single/multi/text
	Sort     int    `gorm:"default:0" json:"sort"`
	Required bool   `gorm:"default:false" json:"required"`

	Options []SurveyOption `gorm:"foreignKey:QuestionID" json:"options,omitempty"`
}

func (SurveyQuestion) TableName() string {
	return "conference_survey_questions"
}

// SurveyOption for single/multi choice questions
type SurveyOption struct {
	BaseModel
	QuestionID uint64 `gorm:"index" json:"questionId"`
	Label      string `gorm:"size:256" json:"label"`
	Sort       int    `gorm:"default:0" json:"sort"`
	Count      int    `gorm:"default:0" json:"count"` // response count
}

func (SurveyOption) TableName() string {
	return "conference_survey_options"
}

// SurveyAnswer stores a user's answer
type SurveyAnswer struct {
	BaseModel
	SurveyID    uint64     `gorm:"index" json:"surveyId"`
	UserID      uint64     `gorm:"index" json:"userId"`
	QuestionID  uint64     `gorm:"index" json:"questionId"`
	Answer      string     `gorm:"type:text" json:"answer"` // text or JSON array of option IDs
	SubmittedAt *time.Time `json:"submittedAt"`

	Survey   *Survey         `gorm:"foreignKey:SurveyID" json:"survey,omitempty"`
	Question *SurveyQuestion `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
}

func (SurveyAnswer) TableName() string {
	return "conference_survey_answers"
}
