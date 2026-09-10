package service

import (
	"application/internal/models"
	"application/pkg/db"
)

type DashboardService struct{}

func (s *DashboardService) UserStats(userID uint64) map[string]interface{} {
	stats := map[string]interface{}{}

	var total int64
	db.DB.Model(&models.Application{}).Where("user_id = ?", userID).Count(&total)
	stats["totalApplications"] = total

	var submitted int64
	db.DB.Model(&models.Application{}).Where("user_id = ? AND status <> ?", userID, models.AppStatusDraft).Count(&submitted)
	stats["submittedApplications"] = submitted

	var passed int64
	db.DB.Model(&models.Application{}).Where("user_id = ? AND status IN ?", userID,
		[]string{models.AppStatusPassed, models.AppStatusPublished, models.AppStatusCertified}).Count(&passed)
	stats["passedApplications"] = passed

	var certCount int64
	db.DB.Model(&models.Certificate{}).Where("user_id = ?", userID).Count(&certCount)
	stats["certificates"] = certCount

	var unread int64
	db.DB.Model(&models.Notification{}).Where("user_id = ? AND read_at IS NULL", userID).Count(&unread)
	stats["unreadNotifications"] = unread

	return stats
}

func (s *DashboardService) AdminStats() map[string]interface{} {
	stats := map[string]interface{}{}

	var batches int64
	db.DB.Model(&models.ProjectBatch{}).Count(&batches)
	stats["totalBatches"] = batches

	var openBatches int64
	db.DB.Model(&models.ProjectBatch{}).Where("status = ?", models.BatchStatusOpen).Count(&openBatches)
	stats["openBatches"] = openBatches

	var applications int64
	db.DB.Model(&models.Application{}).Count(&applications)
	stats["totalApplications"] = applications

	var pendingPreliminary int64
	db.DB.Model(&models.Application{}).Where("status = ?", models.AppStatusSubmitted).Count(&pendingPreliminary)
	stats["pendingPreliminary"] = pendingPreliminary

	var underReview int64
	db.DB.Model(&models.Application{}).Where("status IN ?",
		[]string{models.AppStatusUnderReview, models.AppStatusReviewed}).Count(&underReview)
	stats["underReview"] = underReview

	var passed int64
	db.DB.Model(&models.Application{}).Where("status IN ?",
		[]string{models.AppStatusPassed, models.AppStatusPublished, models.AppStatusCertified}).Count(&passed)
	stats["passed"] = passed

	var users int64
	db.DB.Model(&models.User{}).Count(&users)
	stats["totalUsers"] = users

	var certificates int64
	db.DB.Model(&models.Certificate{}).Count(&certificates)
	stats["certificates"] = certificates

	return stats
}

func (s *DashboardService) ReviewerStats(reviewerID uint64) map[string]interface{} {
	stats := map[string]interface{}{}

	var pending int64
	db.DB.Model(&models.ReviewAssignment{}).
		Where("reviewer_id = ? AND status = ?", reviewerID, models.ReviewStatusPending).Count(&pending)
	stats["pendingReviews"] = pending

	var scored int64
	db.DB.Model(&models.ReviewAssignment{}).
		Where("reviewer_id = ? AND status = ?", reviewerID, models.ReviewStatusScored).Count(&scored)
	stats["scoredReviews"] = scored

	return stats
}
