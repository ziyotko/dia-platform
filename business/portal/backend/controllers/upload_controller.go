package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"server/config"
	"server/services"
	"server/utils"

	"github.com/gin-gonic/gin"
)

type UploadController struct{}

// orgCodePattern 机构编码允许的字符集（作为上传目录段使用，必须无法表达路径）。
var orgCodePattern = regexp.MustCompile(`^[a-z0-9_-]{1,32}$`)

// randomHexToken 生成 n 字节随机数的十六进制串（用于文件名去重）；随机源不可用时返回空串。
func randomHexToken(nBytes int) string {
	buf := make([]byte, nBytes)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return hex.EncodeToString(buf)
}

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
	// 注意：需与前端各页面实际传值保持一致（头像 user、广告 ad、友链 link 等）
	allowedDirs := map[string]bool{
		"": true, "article": true, "attachment": true, "video": true,
		"covers": true, "avatars": true, "images": true, "setting": true,
		"user": true, "ad": true, "link": true,
	}
	if !allowedDirs[dir] {
		ctx.JSON(http.StatusOK, utils.Error(1, "非法上传目录"))
		return
	}

	allowedImageExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	}
	// 视频专用扩展名（唯一允许 800MB 的类型）
	allowedVideoExts := map[string]bool{
		".mp4": true,
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

	// 文件大小限制：视频文件 800MB（与前端一致），其余 50MB，防止磁盘耗尽。
	// 必须按「目录 + 扩展名」双重判定：原实现只看 dir==video 就放行 800MB，
	// 而该分支的扩展名白名单同时包含 zip/rar/pdf/doc 等非视频类型，
	// 等于给任意已认证账号开放了 800MB 的任意文件上传。
	maxSize := int64(50 << 20)
	if dir == "video" && allowedVideoExts[ext] {
		maxSize = 800 << 20
	}
	if file.Size > maxSize {
		ctx.JSON(http.StatusOK, utils.Error(1, "文件大小超出限制"))
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
		// 机构编码会作为上传目录的一段拼进文件系统路径，必须先限定字符集：
		// 若被填成 ".."、"a/b" 等，filepath.Join 会跳出 ./uploads（甚至落到进程当前目录）。
		// 纯展示语义的字段不应具备路径能力，非法值一律忽略并记警告。
		orgCode = strings.ToLower(strings.TrimSpace(settings.OrgCode))
		if orgCode != "" && !orgCodePattern.MatchString(orgCode) {
			utils.Logger.Warnf("机构编码含非法字符，上传时已忽略该目录段: %q", settings.OrgCode)
			orgCode = ""
		}
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

	// 文件名 = 纳秒时间戳 + 随机串：同一纳秒内并发上传（或系统时钟回拨）时仅靠时间戳会重名，
	// 而 SaveUploadedFile 会直接创建/截断同名文件，导致已引用到文章的 URL 指向他人内容。
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), randomHexToken(8), ext)
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
