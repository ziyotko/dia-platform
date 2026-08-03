package service

import (
	"encoding/json"
	"errors"
	"time"

	"conference/internal/models"
	"conference/pkg/db"

	"gorm.io/gorm"
)

type SurveyService struct{}

func (s *SurveyService) Create(survey *models.Survey, questions []models.SurveyQuestion) error {
	tx := db.DB.Begin()
	if err := tx.Create(survey).Error; err != nil {
		tx.Rollback()
		return errors.New("创建问卷失败")
	}
	for i := range questions {
		questions[i].SurveyID = survey.ID
		questions[i].ID = 0
		if err := tx.Create(&questions[i]).Error; err != nil {
			tx.Rollback()
			return errors.New("创建问题失败")
		}
		// Create options
		for j := range questions[i].Options {
			questions[i].Options[j].QuestionID = questions[i].ID
			questions[i].Options[j].ID = 0
			if err := tx.Create(&questions[i].Options[j]).Error; err != nil {
				tx.Rollback()
				return errors.New("创建选项失败")
			}
		}
	}
	return tx.Commit().Error
}

func (s *SurveyService) Update(id uint64, updates map[string]interface{}) error {
	return db.DB.Model(&models.Survey{}).Where("id = ?", id).Updates(updates).Error
}

func (s *SurveyService) Delete(id uint64) error {
	tx := db.DB.Begin()
	// Delete answers
	tx.Where("survey_id = ?", id).Delete(&models.SurveyAnswer{})
	// Delete questions and options
	var questions []models.SurveyQuestion
	tx.Where("survey_id = ?", id).Find(&questions)
	for _, q := range questions {
		tx.Where("question_id = ?", q.ID).Delete(&models.SurveyOption{})
	}
	tx.Where("survey_id = ?", id).Delete(&models.SurveyQuestion{})
	tx.Delete(&models.Survey{}, id)
	return tx.Commit().Error
}

func (s *SurveyService) GetByID(id uint64) (*models.Survey, error) {
	var survey models.Survey
	if err := db.DB.Preload("Questions", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Options", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort ASC")
		}).Order("sort ASC")
	}).First(&survey, id).Error; err != nil {
		return nil, errors.New("问卷不存在")
	}
	return &survey, nil
}

func (s *SurveyService) List(meetingID uint64, status string, page, size int) ([]models.Survey, int64, error) {
	var list []models.Survey
	var total int64
	query := db.DB.Model(&models.Survey{})
	if meetingID > 0 {
		query = query.Where("meeting_id = ?", meetingID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (s *SurveyService) ListAvailable(userID uint64) ([]models.Survey, error) {
	now := time.Now()
	var list []models.Survey
	err := db.DB.Where("status = ? AND start_time <= ? AND end_time >= ?", "open", now, now).
		Order("start_time ASC").Find(&list).Error
	return list, err
}

func (s *SurveyService) SubmitAnswers(surveyID, userID uint64, answers []struct {
	QuestionID uint64 `json:"questionId"`
	Answer     string `json:"answer"` // text or JSON array of option IDs
}) error {
	// Check already submitted
	var count int64
	db.DB.Model(&models.SurveyAnswer{}).Where("survey_id = ? AND user_id = ?", surveyID, userID).Count(&count)
	if count > 0 {
		return errors.New("您已提交过问卷，每人限填一次")
	}

	now := time.Now()
	for _, a := range answers {
		answer := models.SurveyAnswer{
			SurveyID:    surveyID,
			UserID:      userID,
			QuestionID:  a.QuestionID,
			Answer:      a.Answer,
			SubmittedAt: &now,
		}
		if err := db.DB.Create(&answer).Error; err != nil {
			return errors.New("提交失败")
		}

		// Increment option counts for single/multi choice
		var question models.SurveyQuestion
		if db.DB.First(&question, a.QuestionID).Error == nil {
			if question.Type == models.SurveyQTypeSingle {
				db.DB.Model(&models.SurveyOption{}).Where("id = ?", a.Answer).
					UpdateColumn("count", gorm.Expr("count + ?", 1))
			} else if question.Type == models.SurveyQTypeMulti {
				var optionIDs []uint64
				json.Unmarshal([]byte(a.Answer), &optionIDs)
				for _, oid := range optionIDs {
					db.DB.Model(&models.SurveyOption{}).Where("id = ?", oid).
						UpdateColumn("count", gorm.Expr("count + ?", 1))
				}
			}
		}
	}
	return nil
}

func (s *SurveyService) HasUserSubmitted(surveyID, userID uint64) bool {
	var count int64
	db.DB.Model(&models.SurveyAnswer{}).Where("survey_id = ? AND user_id = ?", surveyID, userID).Count(&count)
	return count > 0
}

func (s *SurveyService) GetResults(surveyID uint64) (map[string]interface{}, error) {
	var survey models.Survey
	if err := db.DB.Preload("Questions", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Options", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort ASC")
		}).Order("sort ASC")
	}).First(&survey, surveyID).Error; err != nil {
		return nil, errors.New("问卷不存在")
	}

	var totalRespondents int64
	db.DB.Model(&models.SurveyAnswer{}).Where("survey_id = ?", surveyID).
		Distinct("user_id").Count(&totalRespondents)

	return map[string]interface{}{
		"survey":            survey,
		"total_respondents": totalRespondents,
	}, nil
}

func (s *SurveyService) GetAnswerDetails(surveyID uint64, page, size int) ([]models.SurveyAnswer, int64, error) {
	var list []models.SurveyAnswer
	var total int64
	query := db.DB.Model(&models.SurveyAnswer{}).Preload("Question").Where("survey_id = ?", surveyID)
	query.Count(&total)
	err := query.Order("user_id, question_id ASC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}
