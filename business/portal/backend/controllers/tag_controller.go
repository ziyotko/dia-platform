package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/services"
	"server/utils"
)

type TagController struct {
	tagService *services.TagService
}

func NewTagController() *TagController {
	return &TagController{
		tagService: &services.TagService{},
	}
}

func (c *TagController) GetTags(ctx *gin.Context) {
	name := ctx.Query("name")
	statusStr := ctx.Query("status")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("pageSize", "10")

	status := -1
	if statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			status = s
		}
	}
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize < 1 {
		pageSize = 10
	}

	tags, total, err := c.tagService.GetTags(name, status, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取标签列表失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取标签列表成功", utils.PageData(tags, total, page, pageSize)))
}

func (c *TagController) GetAllTags(ctx *gin.Context) {
	tags, err := c.tagService.GetAllTags()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取标签列表失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取标签列表成功", tags))
}

func (c *TagController) CreateTag(ctx *gin.Context) {
	var req models.Tag
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}
	err := c.tagService.CreateTag(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("创建标签失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("创建标签成功", nil))
}

func (c *TagController) UpdateTag(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "标签ID无效"))
		return
	}
	var req models.Tag
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}
	err = c.tagService.UpdateTag(uint(id), &req)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("更新标签失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新标签成功", nil))
}

func (c *TagController) UpdateTagStatus(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "标签ID无效"))
		return
	}
	var req struct {
		Status int `json:"status"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误"))
		return
	}
	err = c.tagService.UpdateTagStatus(uint(id), req.Status)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "更新标签状态失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新标签状态成功", nil))
}

func (c *TagController) DeleteTag(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "标签ID无效"))
		return
	}
	err = c.tagService.DeleteTag(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("删除标签失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("删除标签成功", nil))
}

func (c *TagController) GetTagArticleStats(ctx *gin.Context) {
	stats, err := c.tagService.GetTagArticleStats()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取标签文章统计失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取成功", utils.AllData(stats, int64(len(stats)))))
}
