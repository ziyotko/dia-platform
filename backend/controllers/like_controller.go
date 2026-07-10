package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/utils"
)

type LikeController struct{}

func NewLikeController() *LikeController {
	return &LikeController{}
}

func (c *LikeController) RecordLike(ctx *gin.Context) {
	var req struct {
		ArticleID uint `json:"article_id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误：article_id 必填"))
		return
	}

	like := &models.LikeAnalytics{
		ArticleID: req.ArticleID,
		IP:        ctx.ClientIP(),
	}

	if err := utils.DB.Create(like).Error; err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "记录失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("记录成功", nil))
}
