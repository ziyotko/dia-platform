package service

import (
	"errors"
	"fmt"
	"time"

	"conference/internal/models"
	"conference/pkg/db"

	"github.com/google/uuid"
)

type SignInService struct{}

// GenerateQRCodeToken creates a unique token for QR code sign-in
func (s *SignInService) GenerateQRCodeToken(meetingID, userID, registrationID uint64) (string, error) {
	token := uuid.New().String()
	existing := models.SignIn{
		MeetingID:      meetingID,
		UserID:         userID,
		RegistrationID: registrationID,
		QRCodeToken:    token,
	}
	if err := db.DB.Create(&existing).Error; err != nil {
		return "", errors.New("生成签到码失败")
	}
	return token, nil
}

// SignInByQRCode handles QR code scan sign-in
func (s *SignInService) SignInByQRCode(token string) error {
	var signIn models.SignIn
	if err := db.DB.Where("qr_code_token = ?", token).First(&signIn).Error; err != nil {
		return errors.New("签到码无效")
	}
	if signIn.SignInTime != nil {
		return errors.New("已签到")
	}
	now := time.Now()
	return db.DB.Model(&signIn).Updates(map[string]interface{}{
		"sign_in_time": &now,
		"method":       "qrcode",
	}).Error
}

// SignInOnline handles automatic sign-in when entering live stream
func (s *SignInService) SignInOnline(meetingID, userID uint64) error {
	// Find the user's registration for this meeting
	var reg models.Registration
	if err := db.DB.Where("meeting_id = ? AND user_id = ? AND status = ?",
		meetingID, userID, models.RegStatusApproved).First(&reg).Error; err != nil {
		return errors.New("未报名该会议或报名未通过")
	}

	// Check if already signed in
	var existing models.SignIn
	err := db.DB.Where("meeting_id = ? AND user_id = ? AND sign_in_time IS NOT NULL",
		meetingID, userID).First(&existing).Error
	if err == nil {
		return nil // Already signed in, just update duration
	}

	// Create sign-in record
	now := time.Now()
	signIn := models.SignIn{
		MeetingID:      meetingID,
		UserID:         userID,
		RegistrationID: reg.ID,
		SignInTime:     &now,
		Method:         "online",
		QRCodeToken:    uuid.New().String(),
	}
	return db.DB.Create(&signIn).Error
}

// SignOut records sign-out time and calculates duration
func (s *SignInService) SignOut(meetingID, userID uint64) error {
	var signIn models.SignIn
	if err := db.DB.Where("meeting_id = ? AND user_id = ? AND sign_in_time IS NOT NULL AND sign_out_time IS NULL",
		meetingID, userID).First(&signIn).Error; err != nil {
		return errors.New("签到记录不存在或已签退")
	}
	now := time.Now()
	duration := int(now.Sub(*signIn.SignInTime).Seconds())
	return db.DB.Model(&signIn).Updates(map[string]interface{}{
		"sign_out_time": &now,
		"duration":      duration,
	}).Error
}

// GetUserSignInStatus returns user's sign-in status for a meeting
func (s *SignInService) GetUserSignInStatus(meetingID, userID uint64) (map[string]interface{}, error) {
	var signIn models.SignIn
	err := db.DB.Where("meeting_id = ? AND user_id = ?", meetingID, userID).First(&signIn).Error
	if err != nil {
		return map[string]interface{}{
			"signed_in": false,
			"qr_code":   "",
		}, nil
	}

	signedIn := signIn.SignInTime != nil
	result := map[string]interface{}{
		"signed_in":     signedIn,
		"sign_in_time":  signIn.SignInTime,
		"sign_out_time": signIn.SignOutTime,
		"duration":      signIn.Duration,
		"method":        signIn.Method,
		"qr_code_token": signIn.QRCodeToken,
	}
	return result, nil
}

// GetSignInStats returns statistics for a meeting
func (s *SignInService) GetSignInStats(meetingID uint64) (map[string]interface{}, error) {
	var totalReg, signedIn int64

	db.DB.Model(&models.Registration{}).Where("meeting_id = ? AND status = ?", meetingID, models.RegStatusApproved).Count(&totalReg)
	db.DB.Model(&models.SignIn{}).Where("meeting_id = ? AND sign_in_time IS NOT NULL", meetingID).Count(&signedIn)

	rate := float64(0)
	if totalReg > 0 {
		rate = float64(signedIn) / float64(totalReg) * 100
	}

	return map[string]interface{}{
		"total_registrations": totalReg,
		"signed_in":           signedIn,
		"not_signed_in":       totalReg - signedIn,
		"rate":                fmt.Sprintf("%.1f%%", rate),
	}, nil
}

// ListSignIns returns sign-in records for a meeting
func (s *SignInService) ListSignIns(meetingID uint64, page, size int) ([]models.SignIn, int64, error) {
	var list []models.SignIn
	var total int64
	query := db.DB.Model(&models.SignIn{}).Preload("User").Preload("Registration")
	if meetingID > 0 {
		query = query.Where("meeting_id = ?", meetingID)
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}
