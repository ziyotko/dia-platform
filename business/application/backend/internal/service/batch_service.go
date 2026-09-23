package service

import (
	"errors"
	"time"

	"application/internal/models"
	"application/pkg/db"
)

type BatchService struct{}

// CreateFromPayload builds a batch from a decoded JSON object instead of binding
// straight into the model: the model's *time.Time fields only accept RFC3339,
// while the date pickers send "YYYY-MM-DD HH:mm:ss", which made it impossible to
// create a batch together with its application window.
func (s *BatchService) CreateFromPayload(payload map[string]interface{}) (*models.ProjectBatch, error) {
	clean := pickUpdates(payload, "title", "category_id", "description", "requirements",
		"apply_start", "apply_end", "review_deadline")
	normalizeTimeFields(clean, "apply_start", "apply_end", "review_deadline")
	b := &models.ProjectBatch{
		Title:          stringOf(clean["title"]),
		CategoryID:     toUint64(clean["category_id"]),
		Description:    stringOf(clean["description"]),
		Requirements:   stringOf(clean["requirements"]),
		ApplyStart:     timeOf(clean["apply_start"]),
		ApplyEnd:       timeOf(clean["apply_end"]),
		ReviewDeadline: timeOf(clean["review_deadline"]),
	}
	if err := s.Create(b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *BatchService) Create(b *models.ProjectBatch) error {
	if b.Title == "" {
		return errors.New("请填写批次名称")
	}
	if err := checkCategory(b.CategoryID); err != nil {
		return err
	}
	if b.ApplyStart != nil && b.ApplyEnd != nil && !b.ApplyEnd.After(*b.ApplyStart) {
		return errors.New("申报截止时间必须晚于开始时间")
	}
	b.Status = models.BatchStatusDraft
	return db.DB.Create(b).Error
}

func (s *BatchService) Update(id uint64, updates map[string]interface{}) error {
	clean := pickUpdates(updates, "title", "category_id", "description", "requirements",
		"apply_start", "apply_end", "review_deadline")
	normalizeTimeFields(clean, "apply_start", "apply_end", "review_deadline")
	if len(clean) == 0 {
		return nil
	}
	var b models.ProjectBatch
	if err := db.DB.First(&b, id).Error; err != nil {
		return errors.New("批次不存在")
	}
	// 批次一旦发布，它定义的「这一轮」就固定了：项目类别是所有申报必须匹配的那个
	// （「项目类别须与申报批次一致」），申报期则是已经接受过提交的区间。
	//
	// 锁定规则（按状态）：
	//   draft            全部可改；
	//   open             除 category_id 外可改——延长/缩短申报期是常见需求，不影响已提交的申报；
	//   reviewing/closed 只允许把 apply_end 往将来改（「延长申报期」），配合 Reopen 重开申报。
	// 只对**真正变化**的字段判断：编辑弹窗会回传整份表单，原样回传的字段不算修改。
	dropUnchanged(b, clean)
	if b.Status != models.BatchStatusDraft {
		if _, ok := clean["category_id"]; ok {
			return errors.New("批次已发布，不能修改项目类别")
		}
		if b.Status != models.BatchStatusOpen {
			if _, ok := clean["apply_start"]; ok {
				return errors.New("批次已进入评审/结束，不能修改申报开始时间")
			}
			if _, ok := clean["apply_end"]; ok {
				end := mergeTime(b.ApplyEnd, clean, "apply_end")
				if end == nil || !end.After(time.Now()) {
					return errors.New("批次已进入评审/结束，申报截止时间只能延长到将来")
				}
			}
		}
	}
	if raw, ok := clean["title"]; ok {
		title, _ := raw.(string)
		if title == "" {
			return errors.New("请填写批次名称")
		}
	}
	if raw, ok := clean["category_id"]; ok {
		if err := checkCategory(toUint64(raw)); err != nil {
			return err
		}
	}
	// Validate the merged (not the incoming) window so a batch can never be
	// stored with a deadline that is before its start date.
	start := mergeTime(b.ApplyStart, clean, "apply_start")
	end := mergeTime(b.ApplyEnd, clean, "apply_end")
	if start != nil && end != nil && !end.After(*start) {
		return errors.New("申报截止时间必须晚于开始时间")
	}
	// 条件更新：读-判-写之间批次可能已被发布/结束，否则会绕过上面的字段锁定规则
	res := db.DB.Model(&models.ProjectBatch{}).
		Where("id = ? AND status = ?", id, b.Status).
		Updates(clean)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("批次状态已变更，请刷新后重试")
	}
	return nil
}

// dropUnchanged removes update keys whose value already equals the stored one:
// 编辑弹窗会回传整份表单，只有真正变化的字段才应该触发锁定规则与写入。
func dropUnchanged(b models.ProjectBatch, updates map[string]interface{}) {
	if raw, ok := updates["category_id"]; ok && toUint64(raw) == b.CategoryID {
		delete(updates, "category_id")
	}
	if raw, ok := updates["apply_start"]; ok && sameTime(raw, b.ApplyStart) {
		delete(updates, "apply_start")
	}
	if raw, ok := updates["apply_end"]; ok && sameTime(raw, b.ApplyEnd) {
		delete(updates, "apply_end")
	}
}

// sameTime 比较更新值与库中值：nil 与 nil 视为相同，time.Time 按时刻比较。
func sameTime(raw interface{}, current *time.Time) bool {
	if raw == nil {
		return current == nil
	}
	t, ok := raw.(time.Time)
	if !ok {
		// 解析失败（normalizeTimeFields 不认识的格式）当作“有变化”，交给后续校验报错
		return false
	}
	return current != nil && t.Equal(*current)
}

// mergeTime returns the value a time column will have after applying updates.
// A key absent from updates keeps its current value; an explicit null clears it.
func mergeTime(current *time.Time, updates map[string]interface{}, key string) *time.Time {
	raw, ok := updates[key]
	if !ok {
		return current
	}
	if t, ok := raw.(time.Time); ok {
		return &t
	}
	return nil
}

func (s *BatchService) Delete(id uint64) error {
	var count int64
	db.DB.Model(&models.Application{}).Where("batch_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("该批次下存在申报记录，无法删除")
	}
	var annCount int64
	db.DB.Model(&models.Announcement{}).Where("batch_id = ?", id).Count(&annCount)
	if annCount > 0 {
		return errors.New("该批次下存在结果公示，无法删除")
	}
	return db.DB.Delete(&models.ProjectBatch{}, id).Error
}

// Publish opens a draft batch for applications. A batch can only be published
// once and must have a category plus a complete application window, otherwise it
// would silently stay open for ever.
func (s *BatchService) Publish(id uint64) error {
	var b models.ProjectBatch
	if err := db.DB.First(&b, id).Error; err != nil {
		return errors.New("批次不存在")
	}
	if b.Status != models.BatchStatusDraft {
		return errors.New("仅草稿状态的批次可发布")
	}
	if b.Title == "" {
		return errors.New("请先填写批次名称")
	}
	if err := checkCategory(b.CategoryID); err != nil {
		return err
	}
	if b.ApplyStart == nil || b.ApplyEnd == nil {
		return errors.New("请先设置申报开始和截止时间")
	}
	if !b.ApplyEnd.After(*b.ApplyStart) {
		return errors.New("申报截止时间必须晚于开始时间")
	}
	// 条件更新：重复点击 / 并发只能发布一次
	res := db.DB.Model(&models.ProjectBatch{}).
		Where("id = ? AND status = ?", id, models.BatchStatusDraft).
		Update("status", models.BatchStatusOpen)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("批次状态已变更，请刷新后重试")
	}
	return nil
}

// Close ends a batch that is either accepting applications or under review.
func (s *BatchService) Close(id uint64) error {
	var b models.ProjectBatch
	if err := db.DB.First(&b, id).Error; err != nil {
		return errors.New("批次不存在")
	}
	if b.Status != models.BatchStatusOpen && b.Status != models.BatchStatusReviewing {
		return errors.New("仅申报中或评审中的批次可结束")
	}
	res := db.DB.Model(&models.ProjectBatch{}).
		Where("id = ? AND status IN ?", id, []string{models.BatchStatusOpen, models.BatchStatusReviewing}).
		Update("status", models.BatchStatusClosed)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("批次状态已变更，请刷新后重试")
	}
	return nil
}

// Reopen puts a batch that already moved on (评审中/已结束) back into 申报中, so a
// round that was closed too early can still collect applications.
//
// 前置条件：申报截止时间必须已经在将来（否则重开后 IsOpen 依旧为 false，等于什么都没发生），
// 且该批次还没有任何「已公示/已发证」的申报——那些结果已经对外可见，不能让新申报把
// 公示口径改掉。自动结束（AutoCloseExpired）的批次因此需要先「编辑 → 延长截止时间」。
func (s *BatchService) Reopen(id uint64) error {
	var b models.ProjectBatch
	if err := db.DB.First(&b, id).Error; err != nil {
		return errors.New("批次不存在")
	}
	if b.Status != models.BatchStatusReviewing && b.Status != models.BatchStatusClosed {
		return errors.New("仅评审中或已结束的批次可重开申报")
	}
	if b.ApplyEnd == nil {
		return errors.New("请先设置申报截止时间")
	}
	if !b.ApplyEnd.After(time.Now()) {
		return errors.New("申报截止时间已过，请先在「编辑」里把它延长到将来")
	}
	var published int64
	db.DB.Model(&models.Application{}).
		Where("batch_id = ? AND (published_at IS NOT NULL OR status IN ?)", id,
			[]string{models.AppStatusPublished, models.AppStatusCertified}).
		Count(&published)
	if published > 0 {
		return errors.New("该批次已有公示/发证记录，不能重开申报")
	}
	res := db.DB.Model(&models.ProjectBatch{}).
		Where("id = ? AND status IN ?", id, []string{models.BatchStatusReviewing, models.BatchStatusClosed}).
		Update("status", models.BatchStatusOpen)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("批次状态已变更，请刷新后重试")
	}
	return nil
}

// StartReview moves an open batch into the review stage.
func (s *BatchService) StartReview(id uint64) error {
	var b models.ProjectBatch
	if err := db.DB.First(&b, id).Error; err != nil {
		return errors.New("批次不存在")
	}
	if b.Status != models.BatchStatusOpen {
		return errors.New("仅申报中的批次可进入评审阶段")
	}
	res := db.DB.Model(&models.ProjectBatch{}).
		Where("id = ? AND status = ?", id, models.BatchStatusOpen).
		Update("status", models.BatchStatusReviewing)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("批次状态已变更，请刷新后重试")
	}
	return nil
}

func (s *BatchService) GetByID(id uint64) (*models.ProjectBatch, error) {
	var b models.ProjectBatch
	if err := db.DB.Preload("Category").First(&b, id).Error; err != nil {
		return nil, errors.New("批次不存在")
	}
	return &b, nil
}

func (s *BatchService) List(page, size int, keyword, status string) ([]models.ProjectBatch, int64, error) {
	var list []models.ProjectBatch
	var total int64
	query := db.DB.Model(&models.ProjectBatch{})
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	err := query.Preload("Category").Order("created_at DESC").
		Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

// ListVisible lists the batches an applicant may see: everything that has been
// published already (申报中 / 评审中 / 已结束). Drafts stay internal, but the
// applicant now sees the real stage instead of a column that always reads
// 申报中.
func (s *BatchService) ListVisible(page, size int, keyword string) ([]models.ProjectBatch, int64, error) {
	var list []models.ProjectBatch
	var total int64
	query := db.DB.Model(&models.ProjectBatch{}).Where("status <> ?", models.BatchStatusDraft)
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Preload("Category").Order("created_at DESC").
		Offset((page - 1) * size).Limit(size).Find(&list).Error
	if err != nil {
		return list, total, err
	}
	for i := range list {
		list[i].CanApply = s.IsOpen(&list[i])
	}
	return list, total, nil
}

// IsOpen checks whether the batch is still accepting applications
func (s *BatchService) IsOpen(b *models.ProjectBatch) bool {
	if b.Status != models.BatchStatusOpen {
		return false
	}
	now := time.Now()
	if b.ApplyStart != nil && now.Before(*b.ApplyStart) {
		return false
	}
	if b.ApplyEnd != nil && now.After(*b.ApplyEnd) {
		return false
	}
	return true
}

// IsReviewOpen checks whether the batch still accepts review work.
// A nil review deadline means "no deadline".
func (s *BatchService) IsReviewOpen(b *models.ProjectBatch) bool {
	if b.ReviewDeadline == nil {
		return true
	}
	return !time.Now().After(*b.ReviewDeadline)
}

// AutoCloseExpired moves open batches past their deadline into reviewing
func (s *BatchService) AutoCloseExpired() {
	db.DB.Model(&models.ProjectBatch{}).
		Where("status = ? AND apply_end IS NOT NULL AND apply_end < ?", models.BatchStatusOpen, time.Now()).
		Update("status", models.BatchStatusReviewing)
}
