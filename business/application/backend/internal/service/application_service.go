package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"application/internal/models"
	"application/pkg/db"
	"application/pkg/storage"

	"gorm.io/gorm"
)

type ApplicationService struct{}

// editable reports whether an applicant may still change the submission:
// drafts and preliminary-rejected applications can be edited and re-submitted.
func editable(status string) bool {
	return status == models.AppStatusDraft || status == models.AppStatusPreliminaryRejected
}

func (s *ApplicationService) Create(app *models.Application) error {
	if app.Title == "" || app.BatchID == 0 {
		return errors.New("请填写完整信息")
	}
	if len([]rune(app.Title)) > 100 {
		return errors.New("项目名称不能超过 100 个字符")
	}
	var batch models.ProjectBatch
	if err := db.DB.First(&batch, app.BatchID).Error; err != nil {
		return errors.New("申报批次不存在")
	}
	bs := BatchService{}
	if !bs.IsOpen(&batch) {
		return errors.New("该批次不在申报期内")
	}
	if app.CategoryID == 0 {
		app.CategoryID = batch.CategoryID
	}
	// The batch defines the category of the round: an application may not pick a
	// different one, otherwise it would be reviewed against the wrong rules.
	if batch.CategoryID != 0 && app.CategoryID != batch.CategoryID {
		return errors.New("项目类别须与申报批次一致")
	}
	if err := checkCategory(app.CategoryID); err != nil {
		return err
	}
	app.Status = models.AppStatusDraft
	return db.DB.Create(app).Error
}

func checkCategory(id uint64) error {
	if id == 0 {
		return errors.New("请选择项目类别")
	}
	var count int64
	db.DB.Model(&models.ProjectCategory{}).Where("id = ?", id).Count(&count)
	if count == 0 {
		return errors.New("项目类别不存在")
	}
	return nil
}

func (s *ApplicationService) UpdateDraft(id, userID uint64, updates map[string]interface{}) error {
	clean := pickUpdates(updates, "category_id", "title", "project_brief", "content")
	var app models.Application
	if err := db.DB.First(&app, id).Error; err != nil {
		return errors.New("申报记录不存在")
	}
	if app.UserID != userID {
		return errors.New("无权操作该申报")
	}
	if !editable(app.Status) {
		return errors.New("仅草稿或初审驳回状态可编辑")
	}
	if len(clean) == 0 {
		return nil
	}
	if raw, ok := clean["category_id"]; ok {
		newID := toUint64(raw)
		var batch models.ProjectBatch
		if err := db.DB.First(&batch, app.BatchID).Error; err == nil && batch.CategoryID != 0 && newID != batch.CategoryID {
			return errors.New("项目类别须与申报批次一致")
		}
		if err := checkCategory(newID); err != nil {
			return err
		}
		clean["category_id"] = newID
	}
	if raw, ok := clean["title"]; ok {
		title, _ := raw.(string)
		if title == "" {
			return errors.New("请填写项目名称")
		}
		if len([]rune(title)) > 100 {
			return errors.New("项目名称不能超过 100 个字符")
		}
	}
	return db.DB.Model(&app).Updates(clean).Error
}

// Withdraw returns a submitted application to draft so the applicant can correct
// it before the preliminary review starts.
func (s *ApplicationService) Withdraw(id, userID uint64) error {
	var app models.Application
	if err := db.DB.First(&app, id).Error; err != nil {
		return errors.New("申报记录不存在")
	}
	if app.UserID != userID {
		return errors.New("无权操作该申报")
	}
	if app.Status != models.AppStatusSubmitted {
		return errors.New("仅待初审的申报可撤回")
	}
	return db.DB.Model(&app).Updates(map[string]interface{}{
		"status":       models.AppStatusDraft,
		"submitted_at": nil,
	}).Error
}

func (s *ApplicationService) Submit(id, userID uint64) error {
	var app models.Application
	if err := db.DB.First(&app, id).Error; err != nil {
		return errors.New("申报记录不存在")
	}
	if app.UserID != userID {
		return errors.New("无权操作该申报")
	}
	if !editable(app.Status) {
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
	if err := db.DB.Model(&app).Updates(map[string]interface{}{
		"status":              models.AppStatusSubmitted,
		"submitted_at":        now,
		"preliminary_opinion": "",
	}).Error; err != nil {
		return err
	}
	notifyUser(app.UserID, "申报已提交", "您的项目《"+app.Title+"》已提交，等待初审。", NotifyTypeApplication)
	return nil
}

func (s *ApplicationService) DeleteDraft(id, userID uint64) error {
	var app models.Application
	if err := db.DB.First(&app, id).Error; err != nil {
		return errors.New("申报记录不存在")
	}
	if app.UserID != userID {
		return errors.New("无权操作该申报")
	}
	if !editable(app.Status) {
		return errors.New("仅草稿或初审驳回状态可删除")
	}
	var materials []models.ApplicationMaterial
	db.DB.Where("application_id = ?", id).Find(&materials)
	db.DB.Where("application_id = ?", id).Delete(&models.ApplicationMaterial{})
	if err := db.DB.Delete(&app).Error; err != nil {
		return err
	}
	// The DB rows are gone, so drop the uploaded files too — otherwise uploads/
	// keeps growing with attachments nobody can reach any more.
	for _, m := range materials {
		storage.RemoveByURL(m.FileURL)
	}
	return nil
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
	app.ReviewCount = int64(len(app.Reviews))
	for i := range app.Reviews {
		if app.Reviews[i].Status == models.ReviewStatusScored {
			app.ScoredCount++
		}
	}
	// The certificate is a separate row; attached here so the detail page can
	// offer "颁发证书" / "查看证书" without a second request. A voided
	// certificate is kept in the table but must not look active.
	var cert models.Certificate
	if err := db.DB.Where("application_id = ? AND status <> ?", id, models.CertStatusVoid).
		First(&cert).Error; err == nil {
		app.Certificate = &cert
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
	fillReviewCounts(list)
	return list, total, err
}

// List returns applications for the admin side. `statuses` (comma separated)
// takes precedence over `status` when both are provided.
func (s *ApplicationService) List(page, size int, batchID uint64, status, statuses, keyword string) ([]models.Application, int64, error) {
	var list []models.Application
	var total int64
	query := db.DB.Model(&models.Application{})
	if batchID > 0 {
		query = query.Where("batch_id = ?", batchID)
	}
	if statuses != "" {
		query = query.Where("status IN ?", strings.Split(statuses, ","))
	} else if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("title LIKE ? OR project_brief LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Preload("Batch").Preload("Category").Preload("User").Order("created_at DESC").
		Offset((page - 1) * size).Limit(size).Find(&list).Error
	fillReviewCounts(list)
	return list, total, err
}

// fillReviewCounts attaches the assignment totals of a page of applications in
// a single grouped query (no N+1). Callers use scored_count to decide between
// showing the average and showing "no score yet".
func fillReviewCounts(list []models.Application) {
	if len(list) == 0 {
		return
	}
	ids := make([]uint64, 0, len(list))
	for _, a := range list {
		ids = append(ids, a.ID)
	}
	type aggregate struct {
		ApplicationID uint64
		Total         int64
		Scored        int64
	}
	var rows []aggregate
	db.DB.Model(&models.ReviewAssignment{}).
		Select("application_id, COUNT(*) AS total, SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) AS scored", models.ReviewStatusScored).
		Where("application_id IN ?", ids).
		Group("application_id").
		Scan(&rows)
	byID := make(map[uint64]aggregate, len(rows))
	for _, r := range rows {
		byID[r.ApplicationID] = r
	}
	for i := range list {
		if r, ok := byID[list[i].ID]; ok {
			list[i].ReviewCount = r.Total
			list[i].ScoredCount = r.Scored
		}
	}
}

func (s *ApplicationService) SaveMaterials(applicationID, userID uint64, materials []models.ApplicationMaterial) error {
	var app models.Application
	if err := db.DB.First(&app, applicationID).Error; err != nil {
		return errors.New("申报记录不存在")
	}
	if app.UserID != userID {
		return errors.New("无权操作该申报")
	}
	if !editable(app.Status) {
		return errors.New("仅草稿或初审驳回状态可上传材料")
	}
	for _, m := range materials {
		if m.Name == "" || m.FileURL == "" {
			return errors.New("材料信息不完整，请重新上传")
		}
	}
	var previous []models.ApplicationMaterial
	db.DB.Where("application_id = ?", applicationID).Find(&previous)

	for i := range materials {
		materials[i].ApplicationID = applicationID
		materials[i].ID = 0
	}
	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("application_id = ?", applicationID).Delete(&models.ApplicationMaterial{}).Error; err != nil {
			return err
		}
		if len(materials) == 0 {
			return nil
		}
		return tx.Create(&materials).Error
	}); err != nil {
		return err
	}

	// The caller always sends the complete list, so anything that was stored
	// before and is not in it any more has been removed by the applicant: delete
	// its file as well.
	kept := make(map[string]bool, len(materials))
	for _, m := range materials {
		kept[m.FileURL] = true
	}
	for _, m := range previous {
		if !kept[m.FileURL] {
			storage.RemoveByURL(m.FileURL)
		}
	}
	return nil
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
	status := models.AppStatusUnderReview
	if !pass {
		status = models.AppStatusPreliminaryRejected
	}
	if err := db.DB.Model(&app).Updates(map[string]interface{}{
		"status":              status,
		"preliminary_opinion": opinion,
	}).Error; err != nil {
		return err
	}
	if pass {
		notifyUser(app.UserID, "初审通过", "您的项目《"+app.Title+"》已通过初审，进入专家评审阶段。", NotifyTypeReview)
	} else {
		notifyUser(app.UserID, "初审未通过", "您的项目《"+app.Title+"》初审未通过，可在申报期内修改后重新提交。意见："+opinion, NotifyTypeReview)
	}
	return nil
}

// AssignReviewers assigns reviewers to an application. Existing assignments are
// preserved (scores are never discarded); only reviewers removed from the list
// that have not scored yet are dropped.
func (s *ApplicationService) AssignReviewers(id uint64, reviewerIDs []uint64) error {
	var app models.Application
	if err := db.DB.First(&app, id).Error; err != nil {
		return errors.New("申报记录不存在")
	}
	if app.Status != models.AppStatusUnderReview && app.Status != models.AppStatusReviewed {
		return errors.New("当前状态不可分配评审")
	}

	target := make(map[uint64]bool, len(reviewerIDs))
	for _, rid := range reviewerIDs {
		if rid > 0 {
			target[rid] = true
		}
	}
	if len(target) == 0 {
		// Clearing the whole list is the escape hatch for an application that no
		// expert will ever score, so it is only safe while nobody has scored.
		// Otherwise the scores would silently disappear from the average.
		var scored int64
		db.DB.Model(&models.ReviewAssignment{}).
			Where("application_id = ? AND status = ?", id, models.ReviewStatusScored).Count(&scored)
		if scored > 0 {
			return errors.New("已有专家完成评分，不能清空评审人")
		}
		if err := db.DB.Where("application_id = ?", id).Delete(&models.ReviewAssignment{}).Error; err != nil {
			return err
		}
		s.recalculateScore(id)
		return nil
	}
	// Only enabled reviewer accounts may be assigned.
	var validCount int64
	db.DB.Model(&models.Admin{}).
		Where("id IN ? AND role_code = ? AND status = 1", keysOf(target), models.RoleReviewer).
		Count(&validCount)
	if int(validCount) != len(target) {
		return errors.New("存在无效的评审人，请重新选择")
	}

	var existing []models.ReviewAssignment
	db.DB.Where("application_id = ?", id).Find(&existing)
	assigned := make(map[uint64]bool, len(existing))
	for _, a := range existing {
		assigned[a.ReviewerID] = true
	}

	// Adding a reviewer after the deadline is forbidden, but removing one is
	// always allowed — otherwise a reviewer that never scores would leave the
	// application stuck in 待评审 for ever.
	var added int
	for rid := range target {
		if !assigned[rid] {
			added++
		}
	}
	if added > 0 {
		var batch models.ProjectBatch
		if err := db.DB.First(&batch, app.BatchID).Error; err == nil {
			bs := BatchService{}
			if !bs.IsReviewOpen(&batch) {
				return errors.New("该批次评审已截止，不能再新增评审人")
			}
		}
	}

	kept := make(map[uint64]bool, len(existing))
	for _, a := range existing {
		if target[a.ReviewerID] {
			kept[a.ReviewerID] = true
			continue
		}
		if a.Status == models.ReviewStatusScored {
			name := "该评审人"
			var admin models.Admin
			if err := db.DB.First(&admin, a.ReviewerID).Error; err == nil && admin.RealName != "" {
				name = admin.RealName
			}
			return errors.New(name + "已完成评分，无法移除")
		}
		if err := db.DB.Delete(&models.ReviewAssignment{}, a.ID).Error; err != nil {
			return err
		}
	}
	for rid := range target {
		if kept[rid] {
			continue
		}
		if err := db.DB.Create(&models.ReviewAssignment{
			ApplicationID: id,
			ReviewerID:    rid,
			Status:        models.ReviewStatusPending,
		}).Error; err != nil {
			return err
		}
	}
	s.recalculateScore(id)
	return nil
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
	if assignment.Status == models.ReviewStatusScored {
		return errors.New("该评审已提交，不能重复评分")
	}
	var app models.Application
	if err := db.DB.First(&app, assignment.ApplicationID).Error; err != nil {
		return errors.New("申报记录不存在")
	}
	var batch models.ProjectBatch
	if err := db.DB.First(&batch, app.BatchID).Error; err == nil {
		bs := BatchService{}
		if !bs.IsReviewOpen(&batch) {
			return errors.New("该批次评审已截止")
		}
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

// recalculateScore refreshes the aggregate score of an application and moves it
// between under_review / reviewed as the pending review count changes.
func (s *ApplicationService) recalculateScore(applicationID uint64) {
	var app models.Application
	if err := db.DB.First(&app, applicationID).Error; err != nil {
		return
	}
	if app.Status != models.AppStatusUnderReview && app.Status != models.AppStatusReviewed {
		return
	}

	var scored []models.ReviewAssignment
	db.DB.Where("application_id = ? AND status = ?", applicationID, models.ReviewStatusScored).Find(&scored)
	var sum float64
	for _, a := range scored {
		sum += a.Score
	}
	var avg float64
	if len(scored) > 0 {
		avg = sum / float64(len(scored))
	}

	var pending int64
	db.DB.Model(&models.ReviewAssignment{}).
		Where("application_id = ? AND status = ?", applicationID, models.ReviewStatusPending).Count(&pending)

	updates := map[string]interface{}{"total_score": sum, "avg_score": avg}
	completed := pending == 0 && len(scored) > 0
	if completed {
		updates["status"] = models.AppStatusReviewed
	} else if pending > 0 {
		// adding a reviewer after scoring re-opens the review
		updates["status"] = models.AppStatusUnderReview
	}
	db.DB.Model(&models.Application{}).Where("id = ?", applicationID).Updates(updates)

	if completed && app.Status == models.AppStatusUnderReview {
		notifyUser(app.UserID, "评审已完成", "您的项目《"+app.Title+"》专家评审已完成，等待评审结果确认。", NotifyTypeReview)
	}
}

// Finalize marks a reviewed application as passed or rejected (评审结果).
//
// An application left in 待评审 with nothing pending (no reviewer assigned at
// all, or every assignment removed) can be decided directly: otherwise it would
// be stuck for ever, since reviewed is only reachable through a score.
func (s *ApplicationService) Finalize(id uint64, pass bool, opinion string) error {
	var app models.Application
	if err := db.DB.First(&app, id).Error; err != nil {
		return errors.New("申报记录不存在")
	}
	switch app.Status {
	case models.AppStatusReviewed:
		// normal path
	case models.AppStatusUnderReview:
		var pending int64
		db.DB.Model(&models.ReviewAssignment{}).
			Where("application_id = ? AND status = ?", id, models.ReviewStatusPending).Count(&pending)
		if pending > 0 {
			return fmt.Errorf("还有 %d 位评审人未评分，无法确定结果", pending)
		}
	default:
		return errors.New("当前状态不可确定评审结果")
	}
	status := models.AppStatusRejected
	if pass {
		status = models.AppStatusPassed
	}
	if err := db.DB.Model(&app).Updates(map[string]interface{}{
		"status":        status,
		"final_opinion": opinion,
	}).Error; err != nil {
		return err
	}
	if pass {
		notifyUser(app.UserID, "评审结果", "您的项目《"+app.Title+"》已通过评审。", NotifyTypeResult)
	} else {
		notifyUser(app.UserID, "评审结果", "您的项目《"+app.Title+"》未通过评审。意见："+opinion, NotifyTypeResult)
	}
	return nil
}

// PublishResult publishes the final result (结果公示).
//
// Approved projects move to `published` (the only state that can be certified).
// Rejected projects keep the `rejected` state and only record the publish time,
// so the pass/reject outcome is never lost.
func (s *ApplicationService) PublishResult(id uint64) error {
	var app models.Application
	if err := db.DB.First(&app, id).Error; err != nil {
		return errors.New("申报记录不存在")
	}
	if app.Status != models.AppStatusPassed && app.Status != models.AppStatusRejected {
		return errors.New("仅已出评审结果的申报可公示")
	}
	if app.PublishedAt != nil {
		return errors.New("该申报结果已公示")
	}
	updates := map[string]interface{}{"published_at": time.Now()}
	if app.Status == models.AppStatusPassed {
		updates["status"] = models.AppStatusPublished
	}
	if err := db.DB.Model(&app).Updates(updates).Error; err != nil {
		return err
	}
	notifyUser(app.UserID, "结果已公示", "您的项目《"+app.Title+"》评审结果已公示，可前往「结果公示」查看。", NotifyTypeResult)
	return nil
}

// RevokeResult takes back a decision that is not final yet so the operator can
// correct it. It is deliberately two-step and never touches a certified
// application: the certificate has to be voided first, which keeps the
// passed → published → certified flow intact.
//
//	已公示 (published / rejected+publishedAt) →撤回公示，结果可重新公示
//	已通过 / 未通过 (not published)          → 撤回评审结果，回到「评审完成」
func (s *ApplicationService) RevokeResult(id uint64) (string, error) {
	var app models.Application
	if err := db.DB.First(&app, id).Error; err != nil {
		return "", errors.New("申报记录不存在")
	}

	updates := map[string]interface{}{}
	var message string
	switch {
	case app.Status == models.AppStatusCertified:
		return "", errors.New("该申报已颁发证书，请先在「证书管理」中作废证书")
	case app.Status == models.AppStatusPublished:
		// 通过后的公示会改状态，撤回时必须一起退回去，否则会留下一个没有
		// published_at 的 published 记录。
		updates["status"] = models.AppStatusPassed
		updates["published_at"] = nil
		message = "已撤回公示，可重新公示"
	case app.PublishedAt != nil && (app.Status == models.AppStatusPassed || app.Status == models.AppStatusRejected):
		// 未通过的项目公示只写 published_at，状态保持不变。
		updates["published_at"] = nil
		message = "已撤回公示，可重新公示"
	case app.Status == models.AppStatusPassed || app.Status == models.AppStatusRejected:
		updates["status"] = models.AppStatusReviewed
		updates["final_opinion"] = ""
		message = "已撤回评审结果，可重新确定结果"
	default:
		return "", errors.New("当前状态不可撤回评审结果")
	}
	if err := db.DB.Model(&app).Updates(updates).Error; err != nil {
		return "", err
	}
	notifyUser(app.UserID, "评审结果已撤回", "您的项目《"+app.Title+"》的评审结果已被撤回，请留意后续通知。", NotifyTypeResult)
	return message, nil
}
