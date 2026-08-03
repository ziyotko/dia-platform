package service

import (
	"conference/internal/models"
	"conference/pkg/db"
)

type DashboardService struct{}

func (s *DashboardService) GetAdminStats() (map[string]interface{}, error) {
	// Meeting stats
	var totalMeetings, openMeetings, closedMeetings int64
	db.DB.Model(&models.Meeting{}).Count(&totalMeetings)
	db.DB.Model(&models.Meeting{}).Where("status = ?", models.MeetingStatusOpen).Count(&openMeetings)
	db.DB.Model(&models.Meeting{}).Where("status = ?", models.MeetingStatusClosed).Count(&closedMeetings)

	// Registration stats
	var totalRegs, approvedRegs, pendingRegs int64
	db.DB.Model(&models.Registration{}).Count(&totalRegs)
	db.DB.Model(&models.Registration{}).Where("status = ?", models.RegStatusApproved).Count(&approvedRegs)
	db.DB.Model(&models.Registration{}).Where("status = ?", models.RegStatusPending).Count(&pendingRegs)

	// Sign-in stats
	var totalSignIns int64
	db.DB.Model(&models.SignIn{}).Where("sign_in_time IS NOT NULL").Count(&totalSignIns)
	signInRate := float64(0)
	if approvedRegs > 0 {
		signInRate = float64(totalSignIns) / float64(approvedRegs) * 100
	}

	// Finance stats
	var totalIncome, totalRefunded float64
	var paidOrders int64
	db.DB.Model(&models.Order{}).Where("status = ?", models.PayStatusPaid).
		Select("COALESCE(SUM(amount), 0)").Scan(&totalIncome)
	db.DB.Model(&models.Order{}).Where("status = ?", models.PayStatusRefunded).
		Select("COALESCE(SUM(amount), 0)").Scan(&totalRefunded)
	db.DB.Model(&models.Order{}).Where("status IN ?", []string{models.PayStatusPaid, models.PayStatusRefunded}).Count(&paidOrders)

	// Credit stats
	var totalCredits float64
	db.DB.Model(&models.CreditRecord{}).Select("COALESCE(SUM(credits), 0)").Scan(&totalCredits)

	// Survey stats
	var totalSurveys int64
	db.DB.Model(&models.Survey{}).Count(&totalSurveys)

	// User stats
	var totalUsers, validUsers int64
	db.DB.Model(&models.User{}).Count(&totalUsers)
	db.DB.Model(&models.User{}).Where("is_valid = ?", true).Count(&validUsers)

	return map[string]interface{}{
		"total_meetings":         totalMeetings,
		"open_meetings":          openMeetings,
		"closed_meetings":        closedMeetings,
		"total_registrations":    totalRegs,
		"approved_registrations": approvedRegs,
		"pending_registrations":  pendingRegs,
		"total_sign_ins":         totalSignIns,
		"sign_in_rate":           signInRate,
		"total_income":           totalIncome,
		"total_refunded":         totalRefunded,
		"net_income":             totalIncome - totalRefunded,
		"paid_orders":            paidOrders,
		"total_credits":          totalCredits,
		"total_surveys":          totalSurveys,
		"total_users":            totalUsers,
		"valid_users":            validUsers,
	}, nil
}

func (s *DashboardService) GetMemberStats(userID uint64) (map[string]interface{}, error) {
	var totalRegs, approvedRegs int64
	db.DB.Model(&models.Registration{}).Where("user_id = ?", userID).Count(&totalRegs)
	db.DB.Model(&models.Registration{}).Where("user_id = ? AND status = ?", userID, models.RegStatusApproved).Count(&approvedRegs)

	var totalCredits float64
	db.DB.Model(&models.CreditRecord{}).Where("user_id = ?", userID).
		Select("COALESCE(SUM(credits), 0)").Scan(&totalCredits)

	var unreadNotifs int64
	db.DB.Model(&models.NotificationRead{}).Where("user_id = ?", userID).Count(&unreadNotifs)
	// Simplified - in real impl, compare with total notifications targeted at user

	// Upcoming meetings
	var upcomingMeetings int64
	db.DB.Model(&models.Meeting{}).Where("status = ?", models.MeetingStatusOpen).Count(&upcomingMeetings)

	return map[string]interface{}{
		"my_registrations":     totalRegs,
		"approved_count":       approvedRegs,
		"total_credits":        totalCredits,
		"upcoming_meetings":    upcomingMeetings,
		"unread_notifications": unreadNotifs,
	}, nil
}
