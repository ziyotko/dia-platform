package service

import (
	"errors"
	"time"

	"conference/internal/models"
	"conference/pkg/db"

	"gorm.io/gorm"
)

type MeetingService struct{}

func (s *MeetingService) Create(meeting *models.Meeting) error {
	return db.DB.Create(meeting).Error
}

func (s *MeetingService) Update(id uint64, updates map[string]interface{}) error {
	return db.DB.Model(&models.Meeting{}).Where("id = ?", id).Updates(updates).Error
}

func (s *MeetingService) Delete(id uint64) error {
	return db.DB.Delete(&models.Meeting{}, id).Error
}

func (s *MeetingService) GetByID(id uint64) (*models.Meeting, error) {
	var m models.Meeting
	if err := db.DB.Preload("Agendas", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort ASC")
	}).Preload("Guests", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort ASC")
	}).First(&m, id).Error; err != nil {
		return nil, errors.New("会议不存在")
	}
	return &m, nil
}

func (s *MeetingService) List(page, size int, keyword, meetingType, status string, userBranch string) ([]models.Meeting, int64, error) {
	var list []models.Meeting
	var total int64
	query := db.DB.Model(&models.Meeting{})

	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}
	if meetingType != "" {
		query = query.Where("type = ?", meetingType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

// ListAvailable returns meetings available for a user (status=open, not excluded by branch/level)
func (s *MeetingService) ListAvailable(page, size int, keyword string, userBranch, userLevel string) ([]models.Meeting, int64, error) {
	var list []models.Meeting
	var total int64

	query := db.DB.Model(&models.Meeting{}).Where("status = ?", models.MeetingStatusOpen)
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}

	query.Count(&total)
	err := query.Order("start_time ASC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

// CloseMeeting closes a meeting and auto-archives it
func (s *MeetingService) CloseMeeting(id uint64) error {
	return db.DB.Model(&models.Meeting{}).Where("id = ?", id).Update("status", models.MeetingStatusClosed).Error
}

// --- Agendas ---

func (s *MeetingService) SaveAgendas(meetingID uint64, agendas []models.Agenda) error {
	db.DB.Where("meeting_id = ?", meetingID).Delete(&models.Agenda{})
	for i := range agendas {
		agendas[i].MeetingID = meetingID
		agendas[i].ID = 0
	}
	return db.DB.Create(&agendas).Error
}

// --- Guests ---

func (s *MeetingService) SaveGuests(meetingID uint64, guests []models.Guest) error {
	db.DB.Where("meeting_id = ?", meetingID).Delete(&models.Guest{})
	for i := range guests {
		guests[i].MeetingID = meetingID
		guests[i].ID = 0
	}
	return db.DB.Create(&guests).Error
}

// --- Access Control ---

func (s *MeetingService) CheckAccess(meetingID uint64, userBranch, userLevel string, userIsValid bool) (bool, error) {
	if !userIsValid {
		return false, nil
	}
	var m models.Meeting
	if err := db.DB.First(&m, meetingID).Error; err != nil {
		return false, errors.New("会议不存在")
	}
	if m.Status != models.MeetingStatusOpen {
		return false, nil
	}
	// If no access restrictions, allow all
	if len(m.AccessBranches) == 0 && len(m.AccessLevels) == 0 {
		return true, nil
	}
	// Check branch restriction
	if len(m.AccessBranches) > 0 {
		found := false
		for _, b := range m.AccessBranches {
			if b == userBranch {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}
	// Check level restriction
	if len(m.AccessLevels) > 0 {
		found := false
		for _, l := range m.AccessLevels {
			if l == userLevel {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}
	return true, nil
}

// IsRegistrationOpen checks if registration is open for the meeting
func (s *MeetingService) IsRegistrationOpen(meetingID uint64) bool {
	var m models.Meeting
	if err := db.DB.First(&m, meetingID).Error; err != nil {
		return false
	}
	now := time.Now()
	if m.RegStartTime != nil && now.Before(*m.RegStartTime) {
		return false
	}
	if m.RegEndTime != nil && now.After(*m.RegEndTime) {
		return false
	}
	return m.Status == models.MeetingStatusOpen
}
