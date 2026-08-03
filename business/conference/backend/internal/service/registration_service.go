package service

import (
	"errors"
	"time"

	"conference/internal/models"
	"conference/pkg/db"
)

type RegistrationService struct{}

func (s *RegistrationService) Register(meetingID, userID uint64) (*models.Registration, error) {
	// Check meeting exists and is open
	var meeting models.Meeting
	if err := db.DB.First(&meeting, meetingID).Error; err != nil {
		return nil, errors.New("会议不存在")
	}
	if meeting.Status != models.MeetingStatusOpen {
		return nil, errors.New("会议未开放报名")
	}

	// Check already registered
	var existing models.Registration
	err := db.DB.Where("meeting_id = ? AND user_id = ? AND status NOT IN ?",
		meetingID, userID, []string{models.RegStatusCancelled, models.RegStatusRejected}).First(&existing).Error
	if err == nil {
		return nil, errors.New("已报名该会议")
	}

	// Check capacity
	var approvedCount int64
	db.DB.Model(&models.Registration{}).Where("meeting_id = ? AND status = ?", meetingID, models.RegStatusApproved).Count(&approvedCount)

	reg := &models.Registration{
		MeetingID: meetingID,
		UserID:    userID,
	}

	if meeting.Capacity > 0 && int(approvedCount) >= meeting.Capacity {
		if !meeting.AllowWaitlist {
			return nil, errors.New("会议名额已满")
		}
		// Waitlist
		reg.Status = models.RegStatusWaitlist
		// Calculate waitlist position
		var waitlistCount int64
		db.DB.Model(&models.Registration{}).Where("meeting_id = ? AND status = ?", meetingID, models.RegStatusWaitlist).Count(&waitlistCount)
		reg.WaitlistPosition = int(waitlistCount) + 1
	} else if meeting.NeedApproval {
		reg.Status = models.RegStatusPending
	} else {
		reg.Status = models.RegStatusApproved
	}

	if err := db.DB.Create(reg).Error; err != nil {
		return nil, errors.New("报名失败")
	}
	return reg, nil
}

func (s *RegistrationService) Cancel(registrationID, userID uint64) error {
	var reg models.Registration
	if err := db.DB.Where("id = ? AND user_id = ?", registrationID, userID).First(&reg).Error; err != nil {
		return errors.New("报名记录不存在")
	}
	if reg.Status == models.RegStatusCancelled {
		return errors.New("已取消")
	}

	// Check cancel deadline
	var meeting models.Meeting
	db.DB.First(&meeting, reg.MeetingID)
	if meeting.CancelDeadline != nil && time.Now().After(*meeting.CancelDeadline) {
		return errors.New("已超过取消报名截止时间")
	}

	return db.DB.Model(&reg).Update("status", models.RegStatusCancelled).Error
}

func (s *RegistrationService) Approve(id uint64, reviewerID uint64, comment string) error {
	var reg models.Registration
	if err := db.DB.First(&reg, id).Error; err != nil {
		return errors.New("报名记录不存在")
	}
	if reg.Status != models.RegStatusPending && reg.Status != models.RegStatusWaitlist {
		return errors.New("当前状态不可审核")
	}
	now := time.Now()
	return db.DB.Model(&reg).Updates(map[string]interface{}{
		"status":         models.RegStatusApproved,
		"review_comment": comment,
		"reviewed_by":    reviewerID,
		"reviewed_at":    &now,
	}).Error
}

func (s *RegistrationService) Reject(id uint64, reviewerID uint64, comment string) error {
	var reg models.Registration
	if err := db.DB.First(&reg, id).Error; err != nil {
		return errors.New("报名记录不存在")
	}
	now := time.Now()
	return db.DB.Model(&reg).Updates(map[string]interface{}{
		"status":         models.RegStatusRejected,
		"review_comment": comment,
		"reviewed_by":    reviewerID,
		"reviewed_at":    &now,
	}).Error
}

func (s *RegistrationService) PromoteFromWaitlist(registrationID uint64, reviewerID uint64) error {
	var reg models.Registration
	if err := db.DB.First(&reg, registrationID).Error; err != nil {
		return errors.New("报名记录不存在")
	}
	if reg.Status != models.RegStatusWaitlist {
		return errors.New("该报名不在候补队列")
	}

	// Move to approved
	var meeting models.Meeting
	db.DB.First(&meeting, reg.MeetingID)

	now := time.Now()
	updates := map[string]interface{}{
		"status":            models.RegStatusApproved,
		"waitlist_position": 0,
		"reviewed_by":       reviewerID,
		"reviewed_at":       &now,
	}
	if meeting.NeedApproval {
		updates["status"] = models.RegStatusApproved // Already approved by promoting
	}

	return db.DB.Model(&reg).Updates(updates).Error
}

func (s *RegistrationService) List(meetingID uint64, status string, page, size int) ([]models.Registration, int64, error) {
	var list []models.Registration
	var total int64
	query := db.DB.Model(&models.Registration{}).Preload("User").Preload("Meeting")
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

func (s *RegistrationService) ListByUser(userID uint64, status string, page, size int) ([]models.Registration, int64, error) {
	var list []models.Registration
	var total int64
	query := db.DB.Model(&models.Registration{}).Preload("Meeting").Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (s *RegistrationService) GetWaitlist(meetingID uint64) ([]models.Registration, error) {
	var list []models.Registration
	err := db.DB.Preload("User").Where("meeting_id = ? AND status = ?", meetingID, models.RegStatusWaitlist).
		Order("waitlist_position ASC").Find(&list).Error
	return list, err
}

func (s *RegistrationService) GetMyRegistration(meetingID, userID uint64) (*models.Registration, error) {
	var reg models.Registration
	err := db.DB.Where("meeting_id = ? AND user_id = ?", meetingID, userID).First(&reg).Error
	if err != nil {
		return nil, errors.New("未报名")
	}
	return &reg, nil
}
