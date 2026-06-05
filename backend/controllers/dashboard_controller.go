package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"server/services"
	"server/utils"
)

type DashboardController struct {
	adService      *services.AdService
	articleService *services.ArticleService
}

func NewDashboardController() *DashboardController {
	return &DashboardController{
		adService:      &services.AdService{},
		articleService: &services.ArticleService{},
	}
}

func (c *DashboardController) GetStats(ctx *gin.Context) {
	adCount := c.adService.GetAdCount()
	articleCount := c.articleService.GetArticleCount()

	ctx.JSON(http.StatusOK, utils.Success("获取成功", gin.H{
		"adCount":      adCount,
		"articleCount": articleCount,
	}))
}
