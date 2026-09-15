package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Storage interface {
	Put(filename string, reader io.Reader, size int64) (key string, url string, err error)
	Delete(key string) error
}

type LocalStorage struct {
	RootDir string
	BaseURL string
}

func NewLocalStorage(rootDir, baseURL string) (*LocalStorage, error) {
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return nil, err
	}
	return &LocalStorage{RootDir: rootDir, BaseURL: baseURL}, nil
}

func (s *LocalStorage) Put(filename string, reader io.Reader, size int64) (string, string, error) {
	dateDir := time.Now().Format("20060102")
	dir := filepath.Join(s.RootDir, dateDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", err
	}
	key := fmt.Sprintf("%s/%d_%s", dateDir, time.Now().UnixNano(), filename)
	path := filepath.Join(s.RootDir, key)
	file, err := os.Create(path)
	if err != nil {
		return "", "", err
	}
	defer file.Close()
	if _, err := io.Copy(file, reader); err != nil {
		return "", "", err
	}
	url := s.BaseURL + "/" + key
	return key, url, nil
}

func (s *LocalStorage) Delete(key string) error {
	return os.Remove(filepath.Join(s.RootDir, key))
}
