package utils

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
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

// LocalUploadPath 把数据库/上传接口里的路径（如 /uploads/templates/x.pdf、uploads\templates\x.pdf）
// 解析为本地相对路径。仅允许 uploads/ 下的已存在文件，防路径穿越与任意文件读取。
func LocalUploadPath(p string) (string, bool) {
	p = strings.TrimSpace(strings.ReplaceAll(p, "\\", "/"))
	p = strings.TrimPrefix(p, "./")
	p = strings.TrimPrefix(p, "/")
	if p == "" || !strings.HasPrefix(p, "uploads/") || strings.Contains(p, "..") {
		return "", false
	}
	fi, err := os.Stat(p)
	if err != nil || fi.IsDir() {
		return "", false
	}
	return p, true
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
