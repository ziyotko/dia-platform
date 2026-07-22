package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"server/services"
	"server/utils"
)

type StaticPageController struct {
	staticPageService *services.StaticPageService
}

func NewStaticPageController() *StaticPageController {
	return &StaticPageController{
		staticPageService: &services.StaticPageService{},
	}
}

func (c *StaticPageController) GetStaticPages(ctx *gin.Context) {
	pageType := ctx.Query("pageType")
	templateIDStr := ctx.Query("templateId")
	var templateID uint
	if templateIDStr != "" {
		if id, err := strconv.ParseUint(templateIDStr, 10, 32); err == nil {
			templateID = uint(id)
		}
	}
	pages, err := c.staticPageService.GetStaticPages(pageType, templateID)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取静态化页面列表失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取静态化页面列表成功", pages))
}
