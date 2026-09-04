package controllers

import (
	"fmt"
	"net/http"
	"time"

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
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: article_id 必填"))
		return
	}

	// 校验文章存在且已发布，避免伪造/无效 ID 污染统计
	var article models.Article
	if err := utils.DB.Select("id").
		Where("id = ? AND status = ? AND deleted_at IS NULL", req.ArticleID, 1).
		First(&article).Error; err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章不存在或未发布"))
		return
	}

	// 同一 IP 对同一文章在窗口内去重，防止刷量导致统计失真
	ip := utils.RealIP(ctx)
	dedupKey := fmt.Sprintf("analytics:visit:%d:%s", req.ArticleID, ip)
	ok, err := utils.Redis1.SetNX(utils.Ctx, dedupKey, "1", time.Hour).Result()
	if err == nil && !ok {
		ctx.JSON(http.StatusOK, utils.Success("记录成功", nil))
		return
	}

	visit := &models.VisitAnalytics{
		ArticleID: req.ArticleID,
		IP:        ip,
	}

	if err := utils.DB.Create(visit).Error; err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "记录失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("记录成功", nil))
}
