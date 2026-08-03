package service

import (
	"conference/internal/models"
	"conference/pkg/db"
)

type CreditService struct{}

// IssueCredits auto-assigns credits when user signs in
func (s *CreditService) IssueCredits(meetingID, userID uint64) error {
	// Check credit config
	var config models.CreditConfig
	if err := db.DB.Where("meeting_id = ?", meetingID).First(&config).Error; err != nil {
		return nil // No credits configured for this meeting
	}
	if config.Credits <= 0 {
		return nil
	}

	// Check already issued
	var count int64
	db.DB.Model(&models.CreditRecord{}).Where("meeting_id = ? AND user_id = ? AND source = ?",
		meetingID, userID, models.CreditSourceAuto).Count(&count)
	if count > 0 {
		return nil // Already issued
	}

	record := models.CreditRecord{
		UserID:    userID,
		MeetingID: meetingID,
		Credits:   config.Credits,
		Source:    models.CreditSourceAuto,
		Remark:    "签到自动发放",
	}
	return db.DB.Create(&record).Error
}

// ManualAdjustIssues admin manually adjusts credits
func (s *CreditService) ManualAdjust(userID, meetingID, operatorID uint64, credits float64, remark string) error {
	record := models.CreditRecord{
		UserID:     userID,
		MeetingID:  meetingID,
		Credits:    credits,
		Source:     models.CreditSourceManual,
		Remark:     remark,
		OperatorID: operatorID,
	}
	return db.DB.Create(&record).Error
}

// SetMeetingCredits sets credit value for a meeting
func (s *CreditService) SetMeetingCredits(meetingID uint64, credits float64) error {
	var config models.CreditConfig
	err := db.DB.Where("meeting_id = ?", meetingID).First(&config).Error
	if err != nil {
		config = models.CreditConfig{
			MeetingID: meetingID,
			Credits:   credits,
		}
		return db.DB.Create(&config).Error
	}
	return db.DB.Model(&config).Update("credits", credits).Error
}

// GetUserCredits returns total credits and history for a user
func (s *CreditService) GetUserCredits(userID uint64, page, size int) (float64, []models.CreditRecord, int64, error) {
	var total float64
	db.DB.Model(&models.CreditRecord{}).Where("user_id = ?", userID).
		Select("COALESCE(SUM(credits), 0)").Scan(&total)

	var list []models.CreditRecord
	var count int64
	query := db.DB.Model(&models.CreditRecord{}).Preload("Meeting").Where("user_id = ?", userID)
	query.Count(&count)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return total, list, count, err
}

// ListRecords returns credit records (admin)
func (s *CreditService) ListRecords(meetingID uint64, source string, page, size int) ([]models.CreditRecord, int64, error) {
	var list []models.CreditRecord
	var total int64
	query := db.DB.Model(&models.CreditRecord{}).Preload("User").Preload("Meeting")
	if meetingID > 0 {
		query = query.Where("meeting_id = ?", meetingID)
	}
	if source != "" {
		query = query.Where("source = ?", source)
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

// GetStats returns credit statistics
func (s *CreditService) GetStats() (map[string]interface{}, error) {
	var totalIssued float64
	var totalManual float64
	var totalUsers int64

	db.DB.Model(&models.CreditRecord{}).Where("source = ?", models.CreditSourceAuto).
		Select("COALESCE(SUM(credits), 0)").Scan(&totalIssued)
	db.DB.Model(&models.CreditRecord{}).Where("source = ?", models.CreditSourceManual).
		Select("COALESCE(SUM(credits), 0)").Scan(&totalManual)
	db.DB.Model(&models.CreditRecord{}).Distinct("user_id").Count(&totalUsers)

	return map[string]interface{}{
		"total_auto_issued":   totalIssued,
		"total_manual_issued": totalManual,
		"total_users":         totalUsers,
		"total_credits":       totalIssued + totalManual,
	}, nil
}

// GetMeetingCredits returns credit config for a meeting
func (s *CreditService) GetMeetingCredits(meetingID uint64) (float64, error) {
	var config models.CreditConfig
	if err := db.DB.Where("meeting_id = ?", meetingID).First(&config).Error; err != nil {
		return 0, nil
	}
	return config.Credits, nil
}
