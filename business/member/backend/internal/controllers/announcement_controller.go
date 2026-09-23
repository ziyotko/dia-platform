package controllers

import (
	"member/internal/service"
	"member/pkg/response"

	"github.com/gin-gonic/gin"
)

type AnnouncementController struct {
	announceService service.AnnouncementService
}

// Public endpoints

func (ctrl *AnnouncementController) GetPublishedAnnouncements(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 10)
	keyword := c.Query("keyword")
	aType := c.Query("type")

	list, total, err := ctrl.announceService.GetPublishedAnnouncements(page, size, keyword, aType)
	if err != nil {
		response.ServerErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

func (ctrl *AnnouncementController) GetAnnouncement(c *gin.Context) {
	id := parseUint(c.Param("id"))
	ann, err := ctrl.announceService.GetAnnouncement(id)
	if err != nil {
		response.NotFoundFrom(c, err)
		return
	}
	response.Success(c, ann)
}

// Admin endpoints

func (ctrl *AnnouncementController) CreateAnnouncement(c *gin.Context) {
	var req service.CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整信息")
		return
	}
	ann, err := ctrl.announceService.CreateAnnouncement(req)
	if err != nil {
		response.BadRequestFrom(c, err)
		return
	}
	response.SuccessWithMessage(c, "创建成功", ann)
}

func (ctrl *AnnouncementController) UpdateAnnouncement(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req service.UpdateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.announceService.UpdateAnnouncement(id, req); err != nil {
		response.BadRequestFrom(c, err)
		return
	}
	response.SuccessWithMessage(c, "更新成功", nil)
}

func (ctrl *AnnouncementController) DeleteAnnouncement(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.announceService.DeleteAnnouncement(id); err != nil {
		response.BadRequestFrom(c, err)
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}

func (ctrl *AnnouncementController) ListAllAnnouncements(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 10)

	list, total, err := ctrl.announceService.ListAllAnnouncements(page, size)
	if err != nil {
		response.ServerErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  size,
	})
}
