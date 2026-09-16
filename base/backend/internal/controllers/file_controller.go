package controllers

import (
	"fmt"
	"strconv"
	"strings"

	"base/config"
	"base/internal/service"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

type FileController struct {
	service *service.FileService
}

func NewFileController() *FileController {
	return &FileController{service: service.NewFileService()}
}

func (ctl *FileController) Upload(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "请选择文件")
		return
	}
	if !ctl.service.IsAllowedType(fileHeader.Filename) {
		response.FailWithCode(c, response.CodeBadRequest, "不支持的文件类型")
		return
	}
	// 大小上限来自配置（server.max_upload_mb）：先按 header 快速拒绝，存储层再按流式兜底
	maxMB := config.Cfg.Server.MaxUploadMB
	if maxMB > 0 && fileHeader.Size > int64(maxMB)*1024*1024 {
		response.FailWithCode(c, response.CodeBadRequest, fmt.Sprintf("文件大小不能超过 %d MB", maxMB))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	defer file.Close()

	tenantID := c.GetUint64("tenantID")
	userID := c.GetUint64("userID")
	uploaded, err := ctl.service.Upload(tenantID, userID, fileHeader.Filename, file, fileHeader.Size)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, uploaded)
}

func (ctl *FileController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	list, total, err := ctl.service.List(c.GetUint64("tenantID"), page, size)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctl *FileController) Delete(c *gin.Context) {
	id := uint64(parseID(c))
	if err := ctl.service.Delete(id, c.GetUint64("tenantID")); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

// Serve 公开读取已上传文件。
// 路由为 /business_base/api/v1/files/*key（key 形如 20260101/1700000000000000000_name.png，含目录分隔符）。
// 路径校验（含 `..` 过滤 + 根目录包含性检查）与绝对路径解析都在 storage 层完成。
func (ctl *FileController) Serve(c *gin.Context) {
	key := strings.TrimPrefix(c.Param("key"), "/")
	if key == "" || strings.Contains(key, "..") {
		response.FailWithCode(c, response.CodeBadRequest, "非法文件路径")
		return
	}
	path, err := ctl.service.ResolvePath(key)
	if err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "非法文件路径")
		return
	}
	c.File(path)
}
