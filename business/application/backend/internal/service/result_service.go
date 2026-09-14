package service

import (
	"errors"
	"fmt"
	"time"

	"application/internal/models"
	"application/pkg/db"

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
	db.DB.Model(&models.Certificate{}).Where("application_id = ?", c.ApplicationID).Count(&count)
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
	db.DB.Model(&models.Certificate{}).Where("cert_no = ?", c.CertNo).Count(&dup)
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
	if raw, ok := clean["cert_no"]; ok {
		no, _ := raw.(string)
		if no == "" {
			return errors.New("请填写证书编号")
		}
		if no != cert.CertNo {
			var dup int64
			db.DB.Model(&models.Certificate{}).Where("cert_no = ? AND id <> ?", no, id).Count(&dup)
			if dup > 0 {
				return errors.New("证书编号已存在")
			}
		}
	}
	return db.DB.Model(&cert).Updates(clean).Error
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
	err := db.DB.Where("user_id = ?", userID).
		Preload("Application", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Batch")
		}).Order("created_at DESC").Find(&list).Error
	return list, err
}
