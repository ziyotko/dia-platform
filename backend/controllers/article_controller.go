package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/services"
	"server/utils"
)

type ArticleController struct {
	articleService *services.ArticleService
	userService    *services.UserService
}

func NewArticleController() *ArticleController {
	return &ArticleController{
		articleService: &services.ArticleService{},
		userService:    &services.UserService{},
	}
}

func (c *ArticleController) GetArticles(ctx *gin.Context) {
	title := ctx.Query("title")
	categoryIDStr := ctx.Query("categoryId")
	statusStr := ctx.Query("status")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("pageSize", "10")

	categoryID := 0
	if categoryIDStr != "" {
		if id, err := strconv.Atoi(categoryIDStr); err == nil {
			categoryID = id
		}
	}
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

	articles, total, err := c.articleService.GetArticles(title, categoryID, status, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取文章列表失败"))
		return
	}

	var list []gin.H
	for _, a := range articles {
		categoryName := ""
		if a.Category.ID > 0 {
			categoryName = a.Category.Name
		}
		tagIds := make([]uint, 0, len(a.Tags))
		for _, t := range a.Tags {
			tagIds = append(tagIds, t.ID)
		}
		list = append(list, gin.H{
			"id":           a.ID,
			"title":        a.Title,
			"categoryId":   a.CategoryID,
			"categoryName": categoryName,
			"summary":      a.Summary,
			"content":      a.Content,
			"status":       a.Status,
			"isTop":        a.IsTop,
			"cover":        a.Cover,
			"author":       a.Author,
			"authorCode":   a.AuthorCode,
			"source":       a.Source,
			"views":        a.Views,
			"tagIds":       tagIds,
			"createTime":   a.CreatedAt.Format("2006-01-02 15:04:05"),
			"updatedAt":    a.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	ctx.JSON(http.StatusOK, utils.Success("获取文章列表成功", gin.H{
		"list":     list,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}))
}

func (c *ArticleController) GetArticleByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章ID无效"))
		return
	}
	article, err := c.articleService.GetArticleByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取文章失败"))
		return
	}

	tagIds := make([]uint, 0, len(article.Tags))
	for _, t := range article.Tags {
		tagIds = append(tagIds, t.ID)
	}

	ctx.JSON(http.StatusOK, utils.Success("获取文章成功", gin.H{
		"id":         article.ID,
		"title":      article.Title,
		"categoryId": article.CategoryID,
		"summary":    article.Summary,
		"content":    article.Content,
		"status":     article.Status,
		"isTop":      article.IsTop,
		"cover":      article.Cover,
		"author":     article.Author,
		"authorCode": article.AuthorCode,
		"source":     article.Source,
		"views":      article.Views,
		"tagIds":     tagIds,
		"createTime": article.CreatedAt.Format("2006-01-02 15:04:05"),
		"updatedAt":  article.UpdatedAt.Format("2006-01-02 15:04:05"),
	}))
}

func (c *ArticleController) CreateArticle(ctx *gin.Context) {
	var req struct {
		models.Article
		TagIds []uint `json:"tagIds"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: "+err.Error()))
		return
	}

	userID := ctx.GetUint("userID")
	user, err := c.userService.GetUserByID(userID)
	if err == nil && user != nil {
		req.Article.Author = user.Username
		req.Article.AuthorCode = strconv.FormatUint(uint64(user.ID), 10)
	}

	err = c.articleService.CreateArticle(&req.Article, req.TagIds)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "创建文章失败: "+err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("创建文章成功", nil))
}

func (c *ArticleController) UpdateArticle(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章ID无效"))
		return
	}
	var req struct {
		models.Article
		TagIds []uint `json:"tagIds"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: "+err.Error()))
		return
	}

	userID := ctx.GetUint("userID")
	user, err := c.userService.GetUserByID(userID)
	if err == nil && user != nil {
		req.Article.Author = user.Username
		req.Article.AuthorCode = strconv.FormatUint(uint64(user.ID), 10)
	}

	err = c.articleService.UpdateArticle(uint(id), &req.Article, req.TagIds)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "更新文章失败: "+err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新文章成功", nil))
}

func (c *ArticleController) UpdateArticleStatus(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章ID无效"))
		return
	}
	var req struct {
		Status int `json:"status"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误"))
		return
	}
	err = c.articleService.UpdateArticleStatus(uint(id), req.Status)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "更新文章状态失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新文章状态成功", nil))
}

func (c *ArticleController) DeleteArticle(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章ID无效"))
		return
	}
	err = c.articleService.DeleteArticle(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "删除文章失败: "+err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("删除文章成功", nil))
}
