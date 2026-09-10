package models

// ProjectCategory is an application category (项目类别)
type ProjectCategory struct {
	BaseModel
	Name        string `gorm:"size:128" json:"name"`
	Description string `gorm:"size:512" json:"description"`
	Sort        int    `gorm:"default:0" json:"sort"`
}

func (ProjectCategory) TableName() string {
	return "application_project_categories"
}
