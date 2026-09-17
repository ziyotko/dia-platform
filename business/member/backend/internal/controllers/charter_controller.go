package controllers

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"member/internal/service"
	"member/pkg/response"
	"member/pkg/utils"

	"github.com/gin-gonic/gin"
)

// charterMaxPdfSize 章程 PDF 大小上限：章程含图表时体积较大，比通用上传（10MB）放宽到 20MB
const charterMaxPdfSize = 20 << 20

// CharterController 协会章程：富文本正文 + PDF 附件，公开可查看、后台可维护。
type CharterController struct {
	charterService service.CharterService
}

// GetCharter 获取协会章程正文与 PDF 附件信息。
// 公开接口（会员端 /charter 页面、服务中心）与后台编辑接口共用，返回结构一致。
func (ctrl *CharterController) GetCharter(c *gin.Context) {
	content, err := ctrl.charterService.GetCharter()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"content": content,
		"file":    ctrl.existingFileMeta(),
	})
}

// SaveCharter 后台保存协会章程正文（富文本 HTML）
func (ctrl *CharterController) SaveCharter(c *gin.Context) {
	var req struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.charterService.SaveCharter(req.Content); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "保存成功", nil)
}

// GetCharterFile 后台查询章程 PDF 附件信息（未上传返回 null）
func (ctrl *CharterController) GetCharterFile(c *gin.Context) {
	response.Success(c, ctrl.existingFileMeta())
}

// UploadCharterFile 后台上传 / 替换章程 PDF
func (ctrl *CharterController) UploadCharterFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择文件")
		return
	}
	if file.Size > charterMaxPdfSize {
		response.BadRequest(c, "文件大小不能超过 20MB")
		return
	}
	if strings.ToLower(filepath.Ext(file.Filename)) != ".pdf" {
		response.BadRequest(c, "仅支持 PDF 文件")
		return
	}
	// 校验文件头，避免改扩展名上传非法文件
	if err := ensurePDF(file); err != nil {
		response.BadRequest(c, "文件不是有效的 PDF")
		return
	}

	// 保留旧文件路径，成功写入新配置后再删除，避免中途失败导致附件丢失
	old, _ := ctrl.charterService.GetCharterFile()

	path, err := utils.SaveUploadedFile(file, "charter")
	if err != nil {
		response.ServerError(c, "文件上传失败")
		return
	}
	meta := service.CharterFileMeta{
		Path:       filepath.ToSlash(path),
		Name:       filepath.Base(file.Filename),
		Size:       file.Size,
		UploadedAt: time.Now().Format(time.RFC3339),
	}
	if err := ctrl.charterService.SaveCharterFile(meta); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	if old != nil && old.Path != "" && old.Path != meta.Path {
		removeUpload(old.Path)
	}
	response.SuccessWithMessage(c, "上传成功", meta)
}

// DeleteCharterFile 后台移除章程 PDF（同时删除磁盘文件）
func (ctrl *CharterController) DeleteCharterFile(c *gin.Context) {
	meta, err := ctrl.charterService.GetCharterFile()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	if err := ctrl.charterService.ClearCharterFile(); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	if meta != nil && meta.Path != "" {
		removeUpload(meta.Path)
	}
	response.SuccessWithMessage(c, "已移除章程 PDF", nil)
}

// DownloadCharter 下载章程 PDF。
// 优先返回后台「协会章程」页面上传的附件，其次回退到历史固定路径（运维手工放置的文件）。
func (ctrl *CharterController) DownloadCharter(c *gin.Context) {
	if meta, err := ctrl.charterService.GetCharterFile(); err == nil && meta != nil && meta.Path != "" {
		if p, ok := utils.LocalUploadPath(meta.Path); ok {
			c.Header("Content-Disposition", "attachment; filename=入会章程.pdf")
			c.File(p)
			return
		}
	}
	// 历史固定路径：uploads/charter/charter.pdf / charter.txt
	if _, err := os.Stat("uploads/charter/charter.pdf"); err == nil {
		c.Header("Content-Disposition", "attachment; filename=入会章程.pdf")
		c.File("uploads/charter/charter.pdf")
		return
	}
	if _, err := os.Stat("uploads/charter/charter.txt"); err == nil {
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.Header("Content-Disposition", "attachment; filename=入会章程.txt")
		c.File("uploads/charter/charter.txt")
		return
	}
	response.NotFound(c, "章程文件暂未上传，请联系管理员")
}

// existingFileMeta 返回章程 PDF 元信息；元信息存在但磁盘文件已被手工删除时返回 nil，
// 避免前端展示「下载」按钮却拿到 404。
func (ctrl *CharterController) existingFileMeta() *service.CharterFileMeta {
	meta, err := ctrl.charterService.GetCharterFile()
	if err != nil || meta == nil || meta.Path == "" {
		return nil
	}
	if _, ok := utils.LocalUploadPath(meta.Path); !ok {
		return nil
	}
	return meta
}

// ensurePDF 读取文件头校验是否为 PDF（%PDF-）
func ensurePDF(file *multipart.FileHeader) error {
	f, err := file.Open()
	if err != nil {
		return err
	}
	defer f.Close()
	head := make([]byte, 5)
	if _, err := io.ReadFull(f, head); err != nil {
		return err
	}
	if string(head) != "%PDF-" {
		return errors.New("invalid pdf header")
	}
	return nil
}

// removeUpload 删除 uploads 下的文件，路径非法或文件不存在时静默跳过
func removeUpload(path string) {
	if p, ok := utils.LocalUploadPath(path); ok {
		_ = os.Remove(p)
	}
}
