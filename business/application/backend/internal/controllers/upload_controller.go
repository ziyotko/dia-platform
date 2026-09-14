package controllers

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"application/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadController handles file uploads (申报材料 / 证书附件)
type UploadController struct{}

var allowedExts = map[string]bool{
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	".ppt": true, ".pptx": true, ".txt": true, ".zip": true, ".rar": true,
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
}

func (ctrl *UploadController) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的文件")
		return
	}

	if file.Size > 50<<20 {
		response.Fail(c, "文件大小不能超过 50MB")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExts[ext] {
		response.Fail(c, "不支持的文件类型")
		return
	}

	dir := "uploads/" + time.Now().Format("20060102")
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	dst := filepath.Join(dir, filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.Fail(c, "文件上传失败")
		return
	}

	response.Ok(c, gin.H{
		"name":     file.Filename,
		"fileUrl":  "/" + filepath.ToSlash(dst),
		"fileType": strings.TrimPrefix(ext, "."),
		"fileSize": file.Size,
	})
}
