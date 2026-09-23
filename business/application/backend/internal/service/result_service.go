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

type ResultService struct{}

// --- Announcements (结果公示) ---

func (s *ResultService) CreateAnnouncement(a *models.Announcement) error {
	if a.Title == "" {
		return errors.New("请填写公示标题")
	}
	if a.BatchID > 0 {
		if err := checkBatch(a.BatchID); err != nil {
			return err
		}
	}
	a.Status = models.AnnouncementStatusDraft
	a.PublishedAt = nil
	return db.DB.Create(a).Error
}

func (s *ResultService) UpdateAnnouncement(id uint64, updates map[string]interface{}) error {
	clean := pickUpdates(updates, "title", "content", "batch_id")
	if len(clean) == 0 {
		return nil
	}
	var a models.Announcement
	if err := db.DB.First(&a, id).Error; err != nil {
		return errors.New("公示不存在")
	}
	if a.Status == models.AnnouncementStatusPublished {
		return errors.New("已发布的公示不可编辑")
	}
	if raw, ok := clean["batch_id"]; ok {
		batchID := toUint64(raw)
		if batchID > 0 {
			if err := checkBatch(batchID); err != nil {
				return err
			}
		}
		clean["batch_id"] = batchID
	}
	return db.DB.Model(&a).Updates(clean).Error
}

func (s *ResultService) DeleteAnnouncement(id uint64) error {
	var a models.Announcement
	if err := db.DB.First(&a, id).Error; err != nil {
		return errors.New("公示不存在")
	}
	if a.Status == models.AnnouncementStatusPublished {
		return errors.New("已发布的公示不可删除")
	}
	return db.DB.Delete(&a).Error
}

// PublishAnnouncement publishes a draft announcement. A published announcement
// is final, so it is published exactly once.
func (s *ResultService) PublishAnnouncement(id uint64) error {
	var a models.Announcement
	if err := db.DB.First(&a, id).Error; err != nil {
		return errors.New("公示不存在")
	}
	if a.Status == models.AnnouncementStatusPublished {
		return errors.New("该公示已发布")
	}
	if a.Title == "" {
		return errors.New("请填写公示标题")
	}
	now := time.Now()
	return db.DB.Model(&a).Updates(map[string]interface{}{
		"status":       models.AnnouncementStatusPublished,
		"published_at": now,
	}).Error
}

// checkBatch reports an error when the referenced batch does not exist.
func checkBatch(id uint64) error {
	if id == 0 {
		return nil
	}
	var count int64
	db.DB.Model(&models.ProjectBatch{}).Where("id = ?", id).Count(&count)
	if count == 0 {
		return errors.New("申报批次不存在")
	}
	return nil
}

func (s *ResultService) ListAnnouncements(page, size int, keyword string, onlyPublished bool) ([]models.Announcement, int64, error) {
	var list []models.Announcement
	var total int64
	query := db.DB.Model(&models.Announcement{})
	if onlyPublished {
		query = query.Where("status = ?", models.AnnouncementStatusPublished)
	}
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Preload("Batch").Order("published_at DESC, created_at DESC").
		Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

// --- Certificates (证书) ---

// IssueCertificate issues a certificate for an approved application.
// Only applications whose result has already been published (已公示) qualify, so
// the documented flow passed → published → certified is always respected.
func (s *ResultService) IssueCertificate(c *models.Certificate) error {
	var app models.Application
	if err := db.DB.First(&app, c.ApplicationID).Error; err != nil {
		return errors.New("申报记录不存在")
	}
	if app.Status != models.AppStatusPublished {
		return errors.New("仅已公示且通过的申报可颁发证书")
	}
	var count int64
	db.DB.Model(&models.Certificate{}).
		Where("application_id = ? AND status <> ?", c.ApplicationID, models.CertStatusVoid).Count(&count)
	if count > 0 {
		return errors.New("该申报已颁发证书")
	}
	if c.Title == "" {
		return errors.New("请填写证书名称")
	}
	if c.CertNo == "" {
		// Generate a stable, readable number: CAAM-<year>-<application id>.
		c.CertNo = fmt.Sprintf("CAAM-%s-%06d", time.Now().Format("2006"), c.ApplicationID)
	}
	var dup int64
	// A voided certificate no longer holds its number, otherwise the generated
	// "CAAM-<year>-<application id>" would make re-issuing after a void
	// impossible for the rest of the year.
	db.DB.Model(&models.Certificate{}).
		Where("cert_no = ? AND status <> ?", c.CertNo, models.CertStatusVoid).Count(&dup)
	if dup > 0 {
		return errors.New("证书编号已存在")
	}
	c.UserID = app.UserID
	c.BatchID = app.BatchID
	c.Status = models.CertStatusIssued
	// Default the holder to the applicant so a certificate is never issued
	// unnamed when the operator leaves the field empty.
	if c.Holder == "" {
		var user models.User
		if err := db.DB.First(&user, app.UserID).Error; err == nil {
			c.Holder = user.RealName
		}
	}
	now := time.Now()
	c.IssuedAt = &now
	if err := db.DB.Create(c).Error; err != nil {
		return err
	}
	if err := db.DB.Model(&app).Update("status", models.AppStatusCertified).Error; err != nil {
		return err
	}
	notifyUser(app.UserID, "证书已颁发", "您的项目《"+app.Title+"》证书已颁发，可在「我的证书」中下载。", NotifyTypeCertificate)
	return nil
}

func (s *ResultService) UpdateCertificate(id uint64, updates map[string]interface{}) error {
	clean := pickUpdates(updates, "cert_no", "title", "holder", "file_url")
	if len(clean) == 0 {
		return nil
	}
	var cert models.Certificate
	if err := db.DB.First(&cert, id).Error; err != nil {
		return errors.New("证书不存在")
	}
	if cert.Status == models.CertStatusVoid {
		return errors.New("证书已作废，不可编辑")
	}
	if raw, ok := clean["cert_no"]; ok {
		no, _ := raw.(string)
		if no == "" {
			return errors.New("请填写证书编号")
		}
		if no != cert.CertNo {
			var dup int64
			db.DB.Model(&models.Certificate{}).
				Where("cert_no = ? AND id <> ? AND status <> ?", no, id, models.CertStatusVoid).Count(&dup)
			if dup > 0 {
				return errors.New("证书编号已存在")
			}
		}
	}
	// Replacing the attachment must not leave the previous file behind.
	if raw, ok := clean["file_url"]; ok {
		newURL, _ := raw.(string)
		if cert.FileURL != "" && newURL != cert.FileURL {
			storage.RemoveByURL(cert.FileURL)
		}
	}
	return db.DB.Model(&cert).Updates(clean).Error
}

// VoidCertificate invalidates an issued certificate. The application returns to
// 已公示 so a corrected certificate can be issued afterwards, and the voided row
// is kept (never deleted) as an audit trail.
func (s *ResultService) VoidCertificate(id uint64, reason string) error {
	var cert models.Certificate
	if err := db.DB.First(&cert, id).Error; err != nil {
		return errors.New("证书不存在")
	}
	if cert.Status == models.CertStatusVoid {
		return errors.New("该证书已作废")
	}
	now := time.Now()
	if err := db.DB.Model(&cert).Updates(map[string]interface{}{
		"status":      models.CertStatusVoid,
		"voided_at":   now,
		"void_reason": reason,
	}).Error; err != nil {
		return err
	}

	var app models.Application
	if err := db.DB.First(&app, cert.ApplicationID).Error; err == nil && app.Status == models.AppStatusCertified {
		db.DB.Model(&app).Update("status", models.AppStatusPublished)
	}
	notifyUser(cert.UserID, "证书已作废", "您的项目《"+cert.Title+"》的证书已作废，如需重新颁发请联系管理方。", NotifyTypeCertificate)
	return nil
}

// PublishedResultBatch 是结果公示里附带的批次信息（只取展示需要的字段）。
type PublishedResultBatch struct {
	ID    uint64 `json:"id"`
	Title string `json:"title"`
}

// PublishedResult 是给申报人看的「逐条结果公示」视图。
// 这里必须显式白名单字段：该接口任意已登录申报人都能访问，若直接返回
// models.Application 并 Preload("User")，会把全体申报人的身份证号、手机号、
// 邮箱一起下发（详见上线前复查报告 P0）。
type PublishedResult struct {
	ID           uint64                `json:"id"`
	Title        string                `json:"title"`
	Status       string                `json:"status"`
	BatchID      uint64                `json:"batchId"`
	Batch        *PublishedResultBatch `json:"batch,omitempty"`
	UserRealName string                `json:"userRealName"`
	PublishedAt  *time.Time            `json:"publishedAt"`
}

// ListPublishedResults returns the per-application results that have been made
// public (逐条结果公示). It is the structured counterpart of the free-text
// Announcement and is what the applicant-facing 结果公示 page shows next to it.
func (s *ResultService) ListPublishedResults(page, size int, batchID uint64, keyword string) ([]PublishedResult, int64, error) {
	var rows []models.Application
	var total int64
	query := db.DB.Model(&models.Application{}).Where("published_at IS NOT NULL")
	if batchID > 0 {
		query = query.Where("batch_id = ?", batchID)
	}
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}
	query.Count(&total)
	// 只取展示需要的列，且预加载也只取必要字段：申报人联系方式/身份证号绝不进入该响应。
	err := query.Select("id", "title", "status", "batch_id", "published_at").
		Preload("Batch", func(tx *gorm.DB) *gorm.DB { return tx.Select("id", "title") }).
		Preload("User", func(tx *gorm.DB) *gorm.DB { return tx.Select("id", "real_name") }).
		Order("published_at DESC").
		Offset((page - 1) * size).Limit(size).Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	list := make([]PublishedResult, 0, len(rows))
	for _, a := range rows {
		item := PublishedResult{
			ID:          a.ID,
			Title:       a.Title,
			Status:      a.Status,
			BatchID:     a.BatchID,
			PublishedAt: a.PublishedAt,
		}
		if a.Batch != nil {
			item.Batch = &PublishedResultBatch{ID: a.Batch.ID, Title: a.Batch.Title}
		}
		if a.User != nil {
			item.UserRealName = a.User.RealName
		}
		list = append(list, item)
	}
	return list, total, nil
}

// BuildAnnouncementContent renders the already-published results of a batch as
// announcement text, so the two 公示 implementations no longer have to be kept
// in sync by hand when a manager writes the notice.
func (s *ResultService) BuildAnnouncementContent(batchID uint64) (string, error) {
	if err := checkBatch(batchID); err != nil {
		return "", err
	}
	if batchID == 0 {
		return "", errors.New("请先选择所属批次")
	}
	var apps []models.Application
	if err := db.DB.Preload("User").Where("batch_id = ? AND published_at IS NOT NULL", batchID).
		Order("id ASC").Find(&apps).Error; err != nil {
		return "", err
	}
	if len(apps) == 0 {
		return "", errors.New("该批次暂无已公示的评审结果")
	}
	var b strings.Builder
	var passed int
	for i, a := range apps {
		name := "-"
		if a.User != nil && a.User.RealName != "" {
			name = a.User.RealName
		}
		outcome := "未通过"
		if a.Status == models.AppStatusPublished || a.Status == models.AppStatusCertified {
			outcome = "通过"
			passed++
		}
		fmt.Fprintf(&b, "%d. %s（申报人：%s）——%s\n", i+1, a.Title, name, outcome)
	}
	header := fmt.Sprintf("本批次共受理并公示 %d 个项目，其中通过 %d 个：\n\n", len(apps), passed)
	return header + b.String(), nil
}

func (s *ResultService) ListCertificates(page, size int, keyword string) ([]models.Certificate, int64, error) {
	var list []models.Certificate
	var total int64
	query := db.DB.Model(&models.Certificate{})
	if keyword != "" {
		query = query.Where("title LIKE ? OR cert_no LIKE ? OR holder LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Preload("Application", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Batch")
	}).Order("created_at DESC").
		Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (s *ResultService) ListUserCertificates(userID uint64) ([]models.Certificate, error) {
	var list []models.Certificate
	// A voided certificate is no longer the applicant's to download.
	err := db.DB.Where("user_id = ? AND status <> ?", userID, models.CertStatusVoid).
		Preload("Application", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Batch")
		}).Order("created_at DESC").Find(&list).Error
	return list, err
}
