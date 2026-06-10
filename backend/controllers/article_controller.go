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
	auditStatusStr := ctx.Query("auditStatus")
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
	auditStatus := -1
	if auditStatusStr != "" {
		if s, err := strconv.Atoi(auditStatusStr); err == nil {
			auditStatus = s
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

	articles, total, err := c.articleService.GetArticles(title, categoryID, status, auditStatus, page, pageSize)
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
		columnIds := make([]uint, 0, len(a.Columns))
		for _, c := range a.Columns {
			columnIds = append(columnIds, c.ID)
		}
		list = append(list, gin.H{
			"id":           a.ID,
			"title":        a.Title,
			"categoryId":   a.CategoryID,
			"categoryName": categoryName,
			"summary":      a.Summary,
			"content":      a.Content,
			"status":       a.Status,
			"auditStatus":  a.AuditStatus,
			"isTop":        a.IsTop,
			"isBold":       a.IsBold,
			"defaultColor": a.DefaultColor,
			"cover":        a.Cover,
			"author":       a.Author,
			"authorCode":   a.AuthorCode,
			"source":       a.Source,
			"columnCount":  len(a.Columns),
			"tagIds":       tagIds,
			"columnIds":    columnIds,
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
	columnIds := make([]uint, 0, len(article.Columns))
	for _, c := range article.Columns {
		columnIds = append(columnIds, c.ID)
	}

	ctx.JSON(http.StatusOK, utils.Success("获取文章成功", gin.H{
		"id":           article.ID,
		"title":        article.Title,
		"categoryId":   article.CategoryID,
		"summary":      article.Summary,
		"content":      article.Content,
		"status":       article.Status,
		"auditStatus":  article.AuditStatus,
		"isTop":        article.IsTop,
		"isBold":       article.IsBold,
		"defaultColor": article.DefaultColor,
		"cover":        article.Cover,
		"author":       article.Author,
		"authorCode":   article.AuthorCode,
		"source":       article.Source,
		"columnCount":  len(article.Columns),
		"tagIds":       tagIds,
		"columnIds":    columnIds,
		"createTime":   article.CreatedAt.Format("2006-01-02 15:04:05"),
		"updatedAt":    article.UpdatedAt.Format("2006-01-02 15:04:05"),
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

func (c *ArticleController) AuditArticle(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章ID无效"))
		return
	}
	var req struct {
		AuditStatus int `json:"auditStatus"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误"))
		return
	}
	if req.AuditStatus == 1 {
		// 提交审核
		err = c.articleService.StartArticleAudit(uint(id))
	} else if req.AuditStatus == 2 {
		// 完成审核
		err = c.articleService.CompleteArticleAudit(uint(id))
	} else {
		err = c.articleService.UpdateAuditStatus(uint(id), req.AuditStatus)
	}
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "审核文章失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("审核文章成功", nil))
}

func (c *ArticleController) GetArticleAuditProgress(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章ID无效"))
		return
	}
	audits, err := c.articleService.GetArticleAuditProgress(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取审核进度失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取审核进度成功", audits))
}

func (c *ArticleController) AdvanceArticleAudit(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章ID无效"))
		return
	}
	var req struct {
		ColumnID uint   `json:"columnId"`
		Remark   string `json:"remark"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误"))
		return
	}
	userID := ctx.GetUint("userID")
	err = c.articleService.AdvanceArticleAudit(uint(id), req.ColumnID, userID, req.Remark)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("推进审核成功", nil))
}

func (c *ArticleController) RejectArticleAudit(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章ID无效"))
		return
	}
	var req struct {
		ColumnID uint   `json:"columnId"`
		Remark   string `json:"remark"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误"))
		return
	}
	userID := ctx.GetUint("userID")
	err = c.articleService.RejectArticleAudit(uint(id), req.ColumnID, userID, req.Remark)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("驳回审核成功", nil))
}

func (c *ArticleController) SetArticleColumns(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章ID无效"))
		return
	}
	var req struct {
		ColumnIds []uint `json:"columnIds"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误"))
		return
	}
	err = c.articleService.SetArticleColumns(uint(id), req.ColumnIds)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "设置文章栏目失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("设置文章栏目成功", nil))
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
