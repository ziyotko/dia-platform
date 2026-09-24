package service

import (
	"errors"
	"io"
	"path/filepath"
	"strings"

	"base/config"
	"base/internal/models"
	"base/pkg/db"
	"base/pkg/permmatch"
	"base/pkg/storage"

	"github.com/sirupsen/logrus"
)

type FileService struct {
	storage storage.Storage
}

// NewFileService 创建文件服务。上传目录与单文件大小上限来自配置
// （server.upload_dir / server.max_upload_mb），存储层会把目录转成绝对路径。
func NewFileService() *FileService {
	dir := config.Cfg.Server.UploadDir
	if dir == "" {
		dir = "./uploads"
	}
	maxSize := int64(config.Cfg.Server.MaxUploadMB) * 1024 * 1024
	s, err := storage.NewLocalStorage(dir, permmatch.APIPrefix()+"/files", maxSize)
	if err != nil {
		// 必须返回零值 FileService：若把 nil 的 *LocalStorage 存进接口字段，
		// s.storage == nil 判定会失效（带类型的 nil），后续调用会直接 panic。
		logrus.WithError(err).Warn("初始化本地存储失败，文件上传功能不可用")
		return &FileService{}
	}
	return &FileService{storage: s}
}

// RemoveStoredFile 按存储 key 删除物理文件（租户级联清理等场景使用）。
// 存储未初始化或 key 为空时静默返回，删不掉只影响磁盘占用。
func (s *FileService) RemoveStoredFile(key string) error {
	if s.storage == nil || key == "" {
		return nil
	}
	return s.storage.Delete(key)
}

// ResolvePath 把存储 key 解析为绝对路径（已校验不会越出上传目录）。
func (s *FileService) ResolvePath(key string) (string, error) {
	if s.storage == nil {
		return "", errors.New("存储未初始化")
	}
	return s.storage.Resolve(key)
}

func (s *FileService) Upload(tenantID, userID uint64, filename string, reader io.Reader, size int64) (*models.UploadedFile, error) {
	if s.storage == nil {
		return nil, errors.New("存储未初始化，请联系管理员")
	}
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
	if s.storage == nil {
		return errors.New("存储未初始化，请联系管理员")
	}
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
