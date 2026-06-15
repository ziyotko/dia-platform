package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	}
	if !allowedExts[ext] {
		ctx.JSON(http.StatusOK, utils.Error(1, "不支持的文件格式，仅允许 jpg/png/gif/webp"))
		return
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

	dir := ctx.DefaultPostForm("dir", "")
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

	if err := ctx.SaveUploadedFile(file, dst); err != nil {
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
		"url": fileURL,
	}))
}
