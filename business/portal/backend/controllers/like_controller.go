package controllers

import (
	"fmt"
	"net/http"
	"time"

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
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: article_id 必填"))
		return
	}

	// 校验文章存在且已发布，避免伪造/无效 ID 污染统计
	var article models.Article
	if err := utils.DB.Select("id").
		Where("id = ? AND status = ?", req.ArticleID, 1).
		First(&article).Error; err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章不存在或未发布"))
		return
	}

	// 同一 IP 对同一文章做点赞去重（较长窗口），防止重复刷赞
	ip := utils.RealIP(ctx)
	dedupKey := fmt.Sprintf("analytics:like:%d:%s", req.ArticleID, ip)
	ok, err := utils.Redis1.SetNX(utils.Ctx, dedupKey, "1", 24*time.Hour).Result()
	if err == nil && !ok {
		ctx.JSON(http.StatusOK, utils.Success("已点赞", nil))
		return
	}

	like := &models.LikeAnalytics{
		ArticleID: req.ArticleID,
		IP:        ip,
	}

	if err := utils.DB.Create(like).Error; err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "记录失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("记录成功", nil))
}
