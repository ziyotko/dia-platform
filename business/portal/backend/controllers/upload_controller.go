package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"server/config"
	"server/services"
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

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		ext = ".png"
	}

	dir := ctx.DefaultPostForm("dir", "")
	// 目录白名单：仅允许已知子目录，阻断路径穿越（如 dir=../../）
	allowedDirs := map[string]bool{
		"": true, "article": true, "attachment": true, "video": true,
		"covers": true, "avatars": true, "images": true,
	}
	if !allowedDirs[dir] {
		ctx.JSON(http.StatusOK, utils.Error(1, "非法上传目录"))
		return
	}

	// 文件大小限制：视频 800MB（与前端一致），其余 50MB，防止磁盘耗尽
	maxSize := int64(50 << 20)
	if dir == "video" {
		maxSize = 800 << 20
	}
	if file.Size > maxSize {
		ctx.JSON(http.StatusOK, utils.Error(1, "文件大小超出限制"))
		return
	}

	allowedImageExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	}
	allowedAttachmentExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
		".ppt": true, ".pptx": true, ".txt": true, ".zip": true, ".rar": true,
		".7z": true, ".mp4": true, ".mp3": true,
	}
	if dir == "article" || dir == "attachment" || dir == "video" {
		if !allowedAttachmentExts[ext] {
			ctx.JSON(http.StatusOK, utils.Error(1, "不支持的文件格式"))
			return
		}
	} else {
		if !allowedImageExts[ext] {
			ctx.JSON(http.StatusOK, utils.Error(1, "不支持的文件格式，仅允许 jpg/png/gif/webp"))
			return
		}
	}

	settingsService := services.SettingsService{}
	settings, err := settingsService.GetSettings()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取系统设置失败"))
		return
	}
	orgCode := ""
	if settings != nil {
		orgCode = strings.ToLower(settings.OrgCode)
	}

	uploadDir := "./uploads"
	if orgCode != "" {
		uploadDir = filepath.Join(uploadDir, orgCode)
	}
	if dir != "" {
		uploadDir = filepath.Join(uploadDir, dir)
	}
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			ctx.JSON(http.StatusOK, utils.Error(1, "创建上传目录失败"))
			return
		}
	} else if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "检查上传目录失败"))
		return
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst := filepath.Join(uploadDir, filename)

	if err := ctx.SaveUploadedFile(file, dst, 0755); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "保存文件失败"))
		return
	}

	fileURL := "/uploads/" + filename
	if orgCode != "" {
		fileURL = "/uploads/" + orgCode + "/" + filename
	}
	if dir != "" {
		if orgCode != "" {
			fileURL = "/uploads/" + orgCode + "/" + dir + "/" + filename
		} else {
			fileURL = "/uploads/" + dir + "/" + filename
		}
	}
	ctx.JSON(http.StatusOK, utils.Success("上传成功", gin.H{
		"url": config.AppConfig.Server.UploadDirPrefix + fileURL,
	}))
}
