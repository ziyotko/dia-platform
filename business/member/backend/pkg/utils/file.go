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
	return dst, nil
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
