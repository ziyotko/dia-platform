package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ErrFileTooLarge 超过存储层大小上限。
var ErrFileTooLarge = errors.New("文件大小超出限制")

type Storage interface {
	Put(filename string, reader io.Reader, size int64) (key string, url string, err error)
	Delete(key string) error
	// Resolve 把存储 key 解析为可读取的绝对路径；key 非法（越出根目录）时返回错误。
	Resolve(key string) (string, error)
}

type LocalStorage struct {
	RootDir string
	BaseURL string
	// MaxSize 单文件大小上限（字节），<=0 表示不限制；写入时按流式校验，避免超大文件写满磁盘
	MaxSize int64
}

func NewLocalStorage(rootDir, baseURL string, maxSize int64) (*LocalStorage, error) {
	// 统一转绝对路径：进程工作目录变化时也能正确定位文件
	abs, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0755); err != nil {
		return nil, err
	}
	return &LocalStorage{RootDir: abs, BaseURL: baseURL, MaxSize: maxSize}, nil
}

func (s *LocalStorage) Put(filename string, reader io.Reader, size int64) (string, string, error) {
	safeName, err := sanitizeFileName(filename)
	if err != nil {
		return "", "", err
	}
	dateDir := time.Now().Format("20060102")
	dir := filepath.Join(s.RootDir, dateDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", err
	}
	key := fmt.Sprintf("%s/%d_%s", dateDir, time.Now().UnixNano(), safeName)
	path := filepath.Join(s.RootDir, filepath.FromSlash(key))
	if !s.contains(path) {
		return "", "", errors.New("非法的文件路径")
	}

	file, err := os.Create(path)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	// LimitReader 多读 1 字节用于判定超限，超限则删除已写入的部分
	src := reader
	if s.MaxSize > 0 {
		src = io.LimitReader(reader, s.MaxSize+1)
	}
	written, err := io.Copy(file, src)
	if err != nil {
		_ = os.Remove(path)
		return "", "", err
	}
	if s.MaxSize > 0 && written > s.MaxSize {
		_ = os.Remove(path)
		return "", "", ErrFileTooLarge
	}
	url := s.BaseURL + "/" + key
	return key, url, nil
}

func (s *LocalStorage) Delete(key string) error {
	path, err := s.Resolve(key)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

// Resolve 校验 key 并返回绝对路径，确保结果始终位于 RootDir 内（防目录穿越）。
func (s *LocalStorage) Resolve(key string) (string, error) {
	clean := filepath.Clean(filepath.Join(s.RootDir, filepath.FromSlash(strings.TrimPrefix(key, "/"))))
	if !s.contains(clean) {
		return "", errors.New("非法的文件路径")
	}
	return clean, nil
}

// contains 判断 target 是否位于 RootDir 内。
func (s *LocalStorage) contains(target string) bool {
	root, err := filepath.Abs(s.RootDir)
	if err != nil {
		return false
	}
	abs, err := filepath.Abs(target)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// sanitizeFileName 只保留基础文件名并过滤危险字符，避免文件名里带目录（`../`、`a/b`）。
func sanitizeFileName(name string) (string, error) {
	name = strings.ReplaceAll(name, "\\", "/")
	name = filepath.Base(name) // 去掉任何目录部分
	name = strings.TrimSpace(name)
	name = strings.Map(func(r rune) rune {
		switch r {
		case '<', '>', ':', '"', '|', '?', '*', 0:
			return '_'
		}
		if r < 32 {
			return -1 // 控制字符直接丢弃
		}
		return r
	}, name)
	name = strings.Trim(name, ". ")
	if name == "" {
		return "", errors.New("文件名不合法")
	}
	if len([]rune(name)) > 120 {
		// 超长文件名：保留扩展名截断，避免超出文件系统限制
		ext := filepath.Ext(name)
		if len(ext) > 20 {
			ext = ""
		}
		keep := 120 - len([]rune(ext))
		if keep < 1 {
			keep = 1
		}
		base := []rune(strings.TrimSuffix(name, ext))
		if len(base) > keep {
			base = base[:keep]
		}
		name = string(base) + ext
	}
	return name, nil
}
