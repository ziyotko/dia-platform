package models

import (
	"time"
)

// Application is a submitted project application (项目申报)
type Application struct {
	BaseModel
	BatchID            uint64     `gorm:"index" json:"batchId"`
	UserID             uint64     `gorm:"index" json:"userId"`
	CategoryID         uint64     `gorm:"index" json:"categoryId"`
	Title              string     `gorm:"size:256" json:"title"`
	ProjectBrief       string     `gorm:"type:text" json:"projectBrief"`
	Content            string     `gorm:"type:text" json:"content"`
	Status             string     `gorm:"size:32;default:draft" json:"status"`
	TotalScore         float64    `gorm:"type:decimal(6,2);default:0" json:"totalScore"`
	AvgScore           float64    `gorm:"type:decimal(6,2);default:0" json:"avgScore"`
	PreliminaryOpinion string     `gorm:"type:text" json:"preliminaryOpinion"`
	FinalOpinion       string     `gorm:"type:text" json:"finalOpinion"`
	SubmittedAt        *time.Time `json:"submittedAt"`
	PublishedAt        *time.Time `json:"publishedAt"`

	Batch     *ProjectBatch         `gorm:"foreignKey:BatchID" json:"batch,omitempty"`
	Category  *ProjectCategory      `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	User      *User                 `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Materials []ApplicationMaterial `gorm:"foreignKey:ApplicationID" json:"materials,omitempty"`
	Reviews   []ReviewAssignment    `gorm:"foreignKey:ApplicationID" json:"reviews,omitempty"`
}

func (Application) TableName() string {
	return "application_applications"
}

// ApplicationMaterial is an uploaded material attached to an application
type ApplicationMaterial struct {
	BaseModel
	ApplicationID uint64 `gorm:"index" json:"applicationId"`
	Name          string `gorm:"size:256" json:"name"`
	FileURL       string `gorm:"size:512" json:"fileUrl"`
	FileType      string `gorm:"size:64" json:"fileType"`
	FileSize      int64  `json:"fileSize"`
}

func (ApplicationMaterial) TableName() string {
	return "application_materials"
}
