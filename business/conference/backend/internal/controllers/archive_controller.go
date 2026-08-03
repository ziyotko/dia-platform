package controllers

import (
	"strconv"

	"conference/internal/middleware"
	"conference/internal/service"
	"conference/pkg/response"

	"github.com/gin-gonic/gin"
)

type ArchiveController struct {
	service service.ArchiveService
}

// --- Admin endpoints ---

func (ctrl *ArchiveController) Archive(c *gin.Context) {
	meetingID := parseUint(c.Param("id"))
	adminID := middleware.GetAdminID(c)

	archive, err := ctrl.service.ArchiveMeeting(meetingID, adminID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, archive)
}

func (ctrl *ArchiveController) List(c *gin.Context) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", "0"))
	meetingType := c.Query("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	list, total, err := ctrl.service.List(year, meetingType, page, size)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctrl *ArchiveController) GetByID(c *gin.Context) {
	id := parseUint(c.Param("id"))
	archive, err := ctrl.service.GetByID(id)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, archive)
}

func (ctrl *ArchiveController) AddMaterial(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req struct {
		ItemType string `json:"itemType" binding:"required"`
		FileName string `json:"fileName" binding:"required"`
		FilePath string `json:"filePath" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.AddMaterial(id, req.ItemType, req.FileName, req.FilePath); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "上传成功", nil)
}
