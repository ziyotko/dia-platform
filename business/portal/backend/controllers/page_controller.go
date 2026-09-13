package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/services"
	"server/utils"
)

type PageController struct {
	pageService *services.PageService
}

func NewPageController() *PageController {
	return &PageController{
		pageService: &services.PageService{},
	}
}

func (c *PageController) GetPages(ctx *gin.Context) {
	pageType := ctx.Query("pageType")
	templateIDStr := ctx.Query("templateId")
	statusStr := ctx.Query("status")
	var templateID uint
	if templateIDStr != "" {
		if id, err := strconv.ParseUint(templateIDStr, 10, 32); err == nil {
			templateID = uint(id)
		}
	}
	var status *int
	if statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			status = &s
		}
	}
	pages, err := c.pageService.GetPages(pageType, templateID, status)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取页面列表失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取页面列表成功", pages))
}

func (c *PageController) CreatePage(ctx *gin.Context) {
	var req models.Page
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}
	err := c.pageService.CreatePage(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("创建页面失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("创建页面成功", nil))
}

func (c *PageController) UpdatePage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "页面ID无效"))
		return
	}
	var req models.Page
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}
	err = c.pageService.UpdatePage(uint(id), &req)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("更新页面失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新页面成功", nil))
}

func (c *PageController) DeletePage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "页面ID无效"))
		return
	}
	err = c.pageService.DeletePage(uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "Cannot delete or update a parent row") {
			ctx.JSON(http.StatusOK, utils.Error(1, "该页面下存在已发布文章的栏目，无法直接删除，请先解除关联"))
			return
		}
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("删除页面失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("删除页面成功", nil))
}
