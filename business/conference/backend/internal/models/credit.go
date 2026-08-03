package models

// CreditConfig defines credits for a meeting
type CreditConfig struct {
	BaseModel
	MeetingID uint64  `gorm:"uniqueIndex" json:"meetingId"`
	Credits   float64 `gorm:"type:decimal(5,1);default:0" json:"credits"`

	Meeting *Meeting `gorm:"foreignKey:MeetingID" json:"meeting,omitempty"`
}

func (CreditConfig) TableName() string {
	return "conference_credit_configs"
}

// CreditRecord stores credit history
type CreditRecord struct {
	BaseModel
	UserID     uint64  `gorm:"index" json:"userId"`
	MeetingID  uint64  `gorm:"index" json:"meetingId"`
	Credits    float64 `gorm:"type:decimal(5,1)" json:"credits"`
	Source     string  `gorm:"size:32" json:"source"` // auto_assign/manual
	Remark     string  `gorm:"size:256" json:"remark"`
	OperatorID uint64  `gorm:"default:0" json:"operatorId"`

	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Meeting *Meeting `gorm:"foreignKey:MeetingID" json:"meeting,omitempty"`
}

func (CreditRecord) TableName() string {
	return "conference_credit_records"
}
