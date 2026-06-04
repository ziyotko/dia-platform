package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"server/utils"

	"github.com/gin-gonic/gin"
)

type UploadController struct{}

func NewUploadController() *UploadController {
	return &UploadController{}
}

func (c *UploadController) UploadFile(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取上传文件失败"))
		return
	}

	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".png"
	}

	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	}
	if !allowedExts[ext] {
		ctx.JSON(http.StatusOK, utils.Error(1, "不支持的文件格式，仅允许 jpg/png/gif/webp"))
		return
	}

	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, os.ModePerm)
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst := filepath.Join(uploadDir, filename)

	if err := ctx.SaveUploadedFile(file, dst); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "保存文件失败"))
		return
	}

	fileURL := "/uploads/" + filename
	ctx.JSON(http.StatusOK, utils.Success("上传成功", gin.H{
		"url": fileURL,
	}))
}
