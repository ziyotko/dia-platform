package service

import (
	"errors"
	"time"

	"application/internal/models"
	"application/pkg/db"

	"gorm.io/gorm"
)

type ApplicationService struct{}

func (s *ApplicationService) Create(app *models.Application) error {
	if app.Title == "" || app.BatchID == 0 {
		return errors.New("请填写完整信息")
	}
	app.Status = models.AppStatusDraft
	return db.DB.Create(app).Error
}

func (s *ApplicationService) UpdateDraft(id, userID uint64, updates map[string]interface{}) error {
	delete(updates, "id")
	delete(updates, "status")
	delete(updates, "user_id")
	delete(updates, "batch_id")
	delete(updates, "total_score")
	delete(updates, "avg_score")
	var app models.Application
	if err := db.DB.First(&app, id).Error; err != nil {
		return errors.New("申报记录不存在")
	}
	if app.UserID != userID {
		return errors.New("无权操作该申报")
	}
	if app.Status != models.AppStatusDraft {
		return errors.New("仅草稿状态可编辑")
	}
	return db.DB.Model(&app).Updates(updates).Error
}

func (s *ApplicationService) Submit(id, userID uint64) error {
	var app models.Application
	if err := db.DB.First(&app, id).Error; err != nil {
		return errors.New("申报记录不存在")
	}
	if app.UserID != userID {
		return errors.New("无权操作该申报")
	}
	if app.Status != models.AppStatusDraft {
		return errors.New("当前状态不可提交")
	}

	var batch models.ProjectBatch
	if err := db.DB.First(&batch, app.BatchID).Error; err != nil {
		return errors.New("批次不存在")
	}
	bs := BatchService{}
	if !bs.IsOpen(&batch) {
		return errors.New("该批次已截止申报")
	}

	var matCount int64
	db.DB.Model(&models.ApplicationMaterial{}).Where("application_id = ?", id).Count(&matCount)
	if matCount == 0 {
		return errors.New("请先上传申报材料")
	}

	now := time.Now()
	return db.DB.Model(&app).Updates(map[string]interface{}{
		"status":       models.AppStatusSubmitted,
		"submitted_at": now,
	}).Error
}

func (s *ApplicationService) DeleteDraft(id, userID uint64) error {
	var app models.Application
	if err := db.DB.First(&app, id).Error; err != nil {
		return errors.New("申报记录不存在")
	}
	if app.UserID != userID {
		return errors.New("无权操作该申报")
	}
	if app.Status != models.AppStatusDraft {
		return errors.New("仅草稿状态可删除")
	}
	db.DB.Where("application_id = ?", id).Delete(&models.ApplicationMaterial{})
	return db.DB.Delete(&app).Error
}

func (s *ApplicationService) GetByID(id uint64) (*models.Application, error) {
	var app models.Application
	if err := db.DB.Preload("Batch").Preload("Category").Preload("User").
		Preload("Materials").
		Preload("Reviews", func(db *gorm.DB) *gorm.DB { return db.Order("id ASC") }).
		Preload("Reviews.Reviewer").
		First(&app, id).Error; err != nil {
		return nil, errors.New("申报记录不存在")
	}
	return &app, nil
}

func (s *ApplicationService) GetForUser(id, userID uint64) (*models.Application, error) {
	app, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	if app.UserID != userID {
		return nil, errors.New("无权查看该申报")
	}
	return app, nil
}

func (s *ApplicationService) ListUser(userID uint64, page, size int, status, keyword string) ([]models.Application, int64, error) {
	var list []models.Application
	var total int64
	query := db.DB.Model(&models.Application{}).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Preload("Batch").Preload("Category").Order("created_at DESC").
		Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (s *ApplicationService) List(page, size int, batchID uint64, status, keyword string) ([]models.Application, int64, error) {
	var list []models.Application
	var total int64
	query := db.DB.Model(&models.Application{})
	if batchID > 0 {
		query = query.Where("batch_id = ?", batchID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("title LIKE ? OR project_brief LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Preload("Batch").Preload("Category").Preload("User").Order("created_at DESC").
		Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (s *ApplicationService) SaveMaterials(applicationID, userID uint64, materials []models.ApplicationMaterial) error {
	var app models.Application
	if err := db.DB.First(&app, applicationID).Error; err != nil {
		return errors.New("申报记录不存在")
	}
	if app.UserID != userID {
		return errors.New("无权操作该申报")
	}
	if app.Status != models.AppStatusDraft {
		return errors.New("仅草稿状态可上传材料")
	}
	db.DB.Where("application_id = ?", applicationID).Delete(&models.ApplicationMaterial{})
	for i := range materials {
		materials[i].ApplicationID = applicationID
		materials[i].ID = 0
	}
	if len(materials) == 0 {
		return nil
	}
	return db.DB.Create(&materials).Error
}

// PreliminaryReview handles online preliminary review (在线初审)
func (s *ApplicationService) PreliminaryReview(id uint64, pass bool, opinion string) error {
	var app models.Application
	if err := db.DB.First(&app, id).Error; err != nil {
		return errors.New("申报记录不存在")
	}
	if app.Status != models.AppStatusSubmitted {
		return errors.New("当前状态不可初审")
	}
	if pass {
		return db.DB.Model(&app).Updates(map[string]interface{}{
			"status":              models.AppStatusUnderReview,
			"preliminary_opinion": opinion,
		}).Error
	}
	return db.DB.Model(&app).Updates(map[string]interface{}{
		"status":              models.AppStatusPreliminaryRejected,
		"preliminary_opinion": opinion,
	}).Error
}

// AssignReviewers assigns reviewers to an application and enters reviewing
func (s *ApplicationService) AssignReviewers(id uint64, reviewerIDs []uint64) error {
	if len(reviewerIDs) == 0 {
		return errors.New("请选择评审人")
	}
	var app models.Application
	if err := db.DB.First(&app, id).Error; err != nil {
		return errors.New("申报记录不存在")
	}
	if app.Status != models.AppStatusUnderReview {
		return errors.New("当前状态不可分配评审")
	}
	// reset previous assignments
	db.DB.Where("application_id = ?", id).Delete(&models.ReviewAssignment{})
	for _, rid := range reviewerIDs {
		if rid == 0 {
			continue
		}
		db.DB.Create(&models.ReviewAssignment{
			ApplicationID: id,
			ReviewerID:    rid,
			Status:        models.ReviewStatusPending,
		})
	}
	return db.DB.Model(&app).Update("status", models.AppStatusUnderReview).Error
}

// SubmitReview lets a reviewer score an assigned application
func (s *ApplicationService) SubmitReview(assignmentID, reviewerID uint64, score float64, comment string) error {
	if score < 0 || score > 100 {
		return errors.New("评分须在 0-100 之间")
	}
	var assignment models.ReviewAssignment
	if err := db.DB.First(&assignment, assignmentID).Error; err != nil {
		return errors.New("评审任务不存在")
	}
	if assignment.ReviewerID != reviewerID {
		return errors.New("无权评审该任务")
	}
	now := time.Now()
	if err := db.DB.Model(&assignment).Updates(map[string]interface{}{
		"status":      models.ReviewStatusScored,
		"score":       score,
		"comment":     comment,
		"reviewed_at": now,
	}).Error; err != nil {
		return err
	}
	s.recalculateScore(assignment.ApplicationID)
	return nil
}

func (s *ApplicationService) recalculateScore(applicationID uint64) {
	var assignments []models.ReviewAssignment
	db.DB.Where("application_id = ? AND status = ?", applicationID, models.ReviewStatusScored).Find(&assignments)
	if len(assignments) == 0 {
		return
	}
	var sum float64
	for _, a := range assignments {
		sum += a.Score
	}
	avg := sum / float64(len(assignments))

	var pending int64
	db.DB.Model(&models.ReviewAssignment{}).
		Where("application_id = ? AND status = ?", applicationID, models.ReviewStatusPending).Count(&pending)

	updates := map[string]interface{}{"total_score": sum, "avg_score": avg}
	if pending == 0 {
		updates["status"] = models.AppStatusReviewed
	}
	db.DB.Model(&models.Application{}).Where("id = ?", applicationID).Updates(updates)
}

// Finalize marks a reviewed application as passed or rejected (评审结果)
func (s *ApplicationService) Finalize(id uint64, pass bool, opinion string) error {
	var app models.Application
	if err := db.DB.First(&app, id).Error; err != nil {
		return errors.New("申报记录不存在")
	}
	if app.Status != models.AppStatusReviewed {
		return errors.New("当前状态不可确定评审结果")
	}
	status := models.AppStatusRejected
	if pass {
		status = models.AppStatusPassed
	}
	return db.DB.Model(&app).Updates(map[string]interface{}{
		"status":        status,
		"final_opinion": opinion,
	}).Error
}

// PublishResult publishes the final result (结果公示)
func (s *ApplicationService) PublishResult(id uint64) error {
	var app models.Application
	if err := db.DB.First(&app, id).Error; err != nil {
		return errors.New("申报记录不存在")
	}
	if app.Status != models.AppStatusPassed && app.Status != models.AppStatusRejected {
		return errors.New("仅已出评审结果的申报可公示")
	}
	return db.DB.Model(&app).Update("status", models.AppStatusPublished).Error
}
