package controllers

import (
	"path/filepath"
	"strings"

	"application/internal/middleware"
	"application/internal/service"
	"application/pkg/response"
	"application/pkg/storage"

	"github.com/gin-gonic/gin"
)

// FileController serves uploaded files through the API with an ownership check.
// 上传目录不再由 main.go 静态托管：文件名是 UUID，但 URL 一旦外泄就永久可读，
// 所以所有预览/下载都走这两个接口（前端用带 token 的 blob 请求 + objectURL）。
type FileController struct {
	service service.FileService
}

// MemberFile 申报人下载自己的材料或自己的证书附件。
func (ctrl *FileController) MemberFile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	ctrl.serve(c, ctrl.service.CanUserRead(userID, c.Query("url")))
}

// AdminFile 管理端下载材料/证书附件：管理人可读全部，评审人仅限分配给自己的申报。
func (ctrl *FileController) AdminFile(c *gin.Context) {
	adminID := middleware.GetAdminID(c)
	roleCode := middleware.GetRoleCode(c)
	ctrl.serve(c, ctrl.service.CanAdminRead(adminID, roleCode, c.Query("url")))
}

func (ctrl *FileController) serve(c *gin.Context, allowed bool) {
	if !allowed {
		// 越权与「文件不属于你」都返回同一个提示，避免泄露文件是否存在
		response.Forbidden(c)
		return
	}
	path, ok := storage.LocalPath(c.Query("url"))
	if !ok {
		response.NotFound(c)
		return
	}
	ext := strings.ToLower(filepath.Ext(path))
	// 压缩包按附件下载，避免浏览器就地展开；同时禁止 MIME 嗅探
	if ext == ".zip" || ext == ".rar" {
		c.Header("Content-Disposition", "attachment; filename=\""+filepath.Base(path)+"\"")
	}
	c.Header("X-Content-Type-Options", "nosniff")
	c.File(path)
}
