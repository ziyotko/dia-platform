package controllers

import (
	"fmt"
	"strconv"

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

func (ctl *FileController) Serve(c *gin.Context) {
	key := c.Param("key")
	path := fmt.Sprintf("./uploads/%s", key)
	c.File(path)
}
