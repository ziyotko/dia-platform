package utils

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

func SaveUploadedFile(file *multipart.FileHeader, subDir string) (string, error) {
	uploadDir := filepath.Join("uploads", subDir, time.Now().Format("2006-01"))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	ext := filepath.Ext(file.Filename)
	newName := uuid.New().String() + ext
	dst := filepath.Join(uploadDir, newName)
	if err := saveMultipartFile(file, dst); err != nil {
		return "", err
	}
	return filepath.ToSlash(dst), nil
}

// SaveBytes 把内存中的数据（如生成的证书 PDF）写入 uploads/<subDir>/YYYY-MM/ 下，
// 返回以 “/” 分隔的相对路径（与 SaveUploadedFile 一致，前端直接拼 / 前缀访问）。
func SaveBytes(subDir, name string, data []byte) (string, error) {
	dir := filepath.Join("uploads", subDir, time.Now().Format("2006-01"))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	dst := filepath.Join(dir, name)
	if err := os.WriteFile(dst, data, 0644); err != nil {
		return "", err
	}
	return filepath.ToSlash(dst), nil
}

func saveMultipartFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}
