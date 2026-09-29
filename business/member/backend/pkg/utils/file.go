package utils

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"member/config"

	"github.com/google/uuid"
)

// ValidateUploadPath 校验「由客户端提交、稍后写入数据库」的上传文件路径。
//
// 只接受本地上传目录下的地址（uploads/... 或 <部署前缀>/uploads/...），拒绝：
//   - 外链 http(s)://（会写入库并在后台渲染成可点击链接）
//   - 协议相对地址 //host/path（前端 fileUrl 会原样输出，浏览器按外站处理 → 钓鱼）
//   - 反斜杠路径与 .. 穿越
//
// 与 LocalUploadPath 的区别：不要求文件已存在（部分流程先写记录后落盘）。
func ValidateUploadPath(p string) error {
	norm := strings.TrimSpace(strings.ReplaceAll(p, "\\", "/"))
	if norm == "" {
		return errors.New("文件路径不能为空")
	}
	if strings.Contains(norm, "..") || strings.Contains(norm, ":") {
		return errors.New("文件路径不合法")
	}
	if strings.HasPrefix(norm, "//") {
		return errors.New("文件路径不合法")
	}
	if !strings.HasPrefix(norm, "uploads/") && !strings.Contains(norm, "/uploads/") {
		return errors.New("文件路径不合法")
	}
	return nil
}

// ValidateOptionalUploadPath 同 ValidateUploadPath，但允许为空（表示未上传）。
func ValidateOptionalUploadPath(p string) error {
	if strings.TrimSpace(p) == "" {
		return nil
	}
	return ValidateUploadPath(p)
}

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

// LocalUploadPath 把数据库/上传接口里的路径（如 /uploads/templates/x.pdf、uploads\templates\x.pdf、
// 带部署前缀的 /business_member/uploads/x.pdf）解析为本地相对路径。
// 仅允许 uploads/ 下的已存在文件，防路径穿越与任意文件读取。
func LocalUploadPath(p string) (string, bool) {
	p = strings.TrimSpace(strings.ReplaceAll(p, "\\", "/"))
	if prefix := strings.TrimSuffix(config.Cfg.Server.UploadDirPrefix, "/"); prefix != "" {
		p = strings.TrimPrefix(p, prefix)
	}
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
