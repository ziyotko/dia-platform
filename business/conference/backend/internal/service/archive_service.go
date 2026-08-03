package service

import (
	"encoding/json"
	"errors"
	"time"

	"conference/internal/models"
	"conference/pkg/db"
)

type ArchiveService struct{}

// ArchiveMeeting gathers all data from a meeting and creates archive
func (s *ArchiveService) ArchiveMeeting(meetingID uint64, operatorID uint64) (*models.Archive, error) {
	var meeting models.Meeting
	if err := db.DB.First(&meeting, meetingID).Error; err != nil {
		return nil, errors.New("会议不存在")
	}

	now := time.Now()
	archive := models.Archive{
		MeetingID:    meeting.ID,
		MeetingTitle: meeting.Title,
		MeetingType:  meeting.Type,
		MeetingYear:  meeting.StartTime.Year(),
		ArchivedAt:   &now,
		OperatorID:   operatorID,
	}
	if err := db.DB.Create(&archive).Error; err != nil {
		return nil, errors.New("创建归档失败")
	}

	// Archive registrations
	var registrations []models.Registration
	db.DB.Preload("User").Where("meeting_id = ?", meetingID).Find(&registrations)
	regJSON, _ := json.Marshal(registrations)
	db.DB.Create(&models.ArchiveItem{
		ArchiveID:   archive.ID,
		ItemType:    models.ArchiveItemRegistration,
		FileName:    "报名名单.json",
		ContentJSON: string(regJSON),
	})

	// Archive sign-ins
	var signIns []models.SignIn
	db.DB.Preload("User").Where("meeting_id = ?", meetingID).Find(&signIns)
	signJSON, _ := json.Marshal(signIns)
	db.DB.Create(&models.ArchiveItem{
		ArchiveID:   archive.ID,
		ItemType:    models.ArchiveItemSignIn,
		FileName:    "签到表.json",
		ContentJSON: string(signJSON),
	})

	// Archive finances
	var orders []models.Order
	db.DB.Where("meeting_id = ?", meetingID).Find(&orders)
	finJSON, _ := json.Marshal(orders)
	db.DB.Create(&models.ArchiveItem{
		ArchiveID:   archive.ID,
		ItemType:    models.ArchiveItemFinance,
		FileName:    "缴费记录.json",
		ContentJSON: string(finJSON),
	})

	// Archive votes
	var votes []models.Vote
	db.DB.Preload("Options").Where("meeting_id = ?", meetingID).Find(&votes)
	voteJSON, _ := json.Marshal(votes)
	db.DB.Create(&models.ArchiveItem{
		ArchiveID:   archive.ID,
		ItemType:    models.ArchiveItemVote,
		FileName:    "投票结果.json",
		ContentJSON: string(voteJSON),
	})

	// Archive surveys
	var surveys []models.Survey
	db.DB.Where("meeting_id = ?", meetingID).Find(&surveys)
	surveyJSON, _ := json.Marshal(surveys)
	db.DB.Create(&models.ArchiveItem{
		ArchiveID:   archive.ID,
		ItemType:    models.ArchiveItemSurvey,
		FileName:    "问卷结果.json",
		ContentJSON: string(surveyJSON),
	})

	// Update meeting status
	db.DB.Model(&meeting).Update("status", models.MeetingStatusArchived)

	return &archive, nil
}

// List returns archives with search
func (s *ArchiveService) List(year int, meetingType string, page, size int) ([]models.Archive, int64, error) {
	var list []models.Archive
	var total int64
	query := db.DB.Model(&models.Archive{})
	if year > 0 {
		query = query.Where("meeting_year = ?", year)
	}
	if meetingType != "" {
		query = query.Where("meeting_type = ?", meetingType)
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

// GetByID returns archive with items
func (s *ArchiveService) GetByID(id uint64) (*models.Archive, error) {
	var archive models.Archive
	if err := db.DB.Preload("Items").First(&archive, id).Error; err != nil {
		return nil, errors.New("归档不存在")
	}
	return &archive, nil
}

// AddMaterial uploads a meeting material or minutes to archive
func (s *ArchiveService) AddMaterial(archiveID uint64, itemType, fileName, filePath string) error {
	item := models.ArchiveItem{
		ArchiveID: archiveID,
		ItemType:  itemType,
		FileName:  fileName,
		FilePath:  filePath,
	}
	return db.DB.Create(&item).Error
}
