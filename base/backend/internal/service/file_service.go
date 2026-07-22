package service

import (
	"io"
	"path/filepath"
	"strings"

	"base/internal/models"
	"base/pkg/db"
	"base/pkg/storage"
)

type FileService struct {
	storage storage.Storage
}

func NewFileService() *FileService {
	s, _ := storage.NewLocalStorage("./uploads", "/base/api/v1/files")
	return &FileService{storage: s}
}

func (s *FileService) Upload(tenantID, userID uint64, filename string, reader io.Reader, size int64) (*models.UploadedFile, error) {
	key, url, err := s.storage.Put(filename, reader, size)
	if err != nil {
		return nil, err
	}
	file := models.UploadedFile{
		TenantID: tenantID,
		UserID:   userID,
		FileName: filename,
		FileKey:  key,
		FileType: filepath.Ext(filename),
		FileSize: size,
		URL:      url,
		Storage:  "local",
	}
	if err := db.DB.Create(&file).Error; err != nil {
		_ = s.storage.Delete(key)
		return nil, err
	}
	return &file, nil
}

func (s *FileService) List(tenantID uint64, page, size int) ([]models.UploadedFile, int64, error) {
	var list []models.UploadedFile
	var total int64
	query := db.DB.Model(&models.UploadedFile{})
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	query.Count(&total)
	offset := (page - 1) * size
	err := query.Order("created_at DESC").Offset(offset).Limit(size).Find(&list).Error
	return list, total, err
}

func (s *FileService) Delete(id uint64, tenantID uint64) error {
	var file models.UploadedFile
	query := db.DB.Where("id = ?", id)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&file).Error; err != nil {
		return err
	}
	_ = s.storage.Delete(file.FileKey)
	return db.DB.Delete(&file).Error
}

func (s *FileService) IsAllowedType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	allowed := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true,
		".txt": true, ".zip": true, ".rar": true, ".mp4": true,
	}
	return allowed[ext]
}
