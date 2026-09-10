package service

import (
	"errors"

	"application/internal/models"
	"application/pkg/db"

	"gorm.io/gorm"
)

type ReviewService struct{}

// ListReviewers returns admins with role reviewer (评审人)
func (s *ReviewService) ListReviewers() ([]models.Admin, error) {
	var list []models.Admin
	err := db.DB.Where("role_code = ? AND status = 1", models.RoleReviewer).Order("id ASC").Find(&list).Error
	return list, err
}

// MyAssignments lists review tasks assigned to a reviewer
func (s *ReviewService) MyAssignments(reviewerID uint64, page, size int, status string) ([]models.ReviewAssignment, int64, error) {
	var list []models.ReviewAssignment
	var total int64
	query := db.DB.Model(&models.ReviewAssignment{}).Where("reviewer_id = ?", reviewerID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	err := query.
		Preload("Application", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Batch").Preload("Category").Preload("Materials")
		}).
		Order("created_at DESC").
		Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

// GetAssignment returns a single assignment with its application
func (s *ReviewService) GetAssignment(id, reviewerID uint64) (*models.ReviewAssignment, error) {
	var a models.ReviewAssignment
	if err := db.DB.Preload("Application", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Batch").Preload("Category").Preload("Materials")
	}).First(&a, id).Error; err != nil {
		return nil, errors.New("评审任务不存在")
	}
	if a.ReviewerID != reviewerID {
		return nil, errors.New("无权查看该任务")
	}
	return &a, nil
}

// ListAssignments lists all assignments of an application (admin view)
func (s *ReviewService) ListAssignments(applicationID uint64) ([]models.ReviewAssignment, error) {
	var list []models.ReviewAssignment
	err := db.DB.Preload("Reviewer").Where("application_id = ?", applicationID).Order("id ASC").Find(&list).Error
	return list, err
}
