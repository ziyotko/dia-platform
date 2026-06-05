package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/utils"
)

type VisitController struct{}

func NewVisitController() *VisitController {
	return &VisitController{}
}

func (c *VisitController) RecordVisit(ctx *gin.Context) {
	var req struct {
		ArticleID uint `json:"article_id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误：article_id 必填"))
		return
	}

	visit := &models.SiteAnalytics{
		ArticleID: req.ArticleID,
		IP:        ctx.ClientIP(),
	}

	if err := utils.DB.Create(visit).Error; err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "记录失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("记录成功", nil))
}
