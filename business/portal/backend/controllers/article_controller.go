package controllers

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/services"
	"server/utils"
)

type ArticleController struct {
	articleService  *services.ArticleService
	userService     *services.UserService
	workflowService *services.WorkflowService
}

func NewArticleController() *ArticleController {
	return &ArticleController{
		articleService:  &services.ArticleService{},
		userService:     &services.UserService{},
		workflowService: &services.WorkflowService{},
	}
}

func (c *ArticleController) GetArticles(ctx *gin.Context) {
	title := ctx.Query("title")
	categoryIDStr := ctx.Query("categoryId")
	tagIDStr := ctx.Query("tagId")
	statusStr := ctx.Query("status")
	auditStatusStr := ctx.Query("auditStatus")
	typeStr := ctx.Query("type")
	author := ctx.Query("author")
	source := ctx.Query("source")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("pageSize", "10")

	categoryID := 0
	if categoryIDStr != "" {
		if id, err := strconv.Atoi(categoryIDStr); err == nil {
			categoryID = id
		}
	}
	tagID := 0
	if tagIDStr != "" {
		if id, err := strconv.Atoi(tagIDStr); err == nil {
			tagID = id
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
	articleType := 0
	if typeStr != "" {
		if t, err := strconv.Atoi(typeStr); err == nil {
			articleType = t
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

	articles, total, err := c.articleService.GetArticles(title, categoryID, tagID, status, auditStatus, articleType, author, source, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取文章列表失败"))
		return
	}

	var list []gin.H
	for _, a := range articles {
		columnIds := make([]uint, 0, len(a.Columns))
		for _, c := range a.Columns {
			columnIds = append(columnIds, c.ID)
		}
		attachments := make([]gin.H, 0, len(a.Attachments))
		for _, att := range a.Attachments {
			attachments = append(attachments, gin.H{
				"id":         att.ID,
				"name":       att.Name,
				"url":        att.URL,
				"size":       att.Size,
				"createTime": att.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
		list = append(list, gin.H{
			"id":           a.ID,
			"title":        a.Title,
			"type":         a.Type,
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
			"publishTime":  formatLocalTime(a.PublishTime),
			"url":          a.URL,
			"columnCount":  len(a.Columns),
			"columnIds":    columnIds,
			"attachments":  attachments,
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

	categoryIds := make([]uint, 0, len(article.Categories))
	categoryNames := make([]string, 0, len(article.Categories))
	for _, c := range article.Categories {
		categoryIds = append(categoryIds, c.ID)
		categoryNames = append(categoryNames, c.Name)
	}
	tagIds := make([]uint, 0, len(article.Tags))
	for _, t := range article.Tags {
		tagIds = append(tagIds, t.ID)
	}
	columnIds := make([]uint, 0, len(article.Columns))
	for _, c := range article.Columns {
		columnIds = append(columnIds, c.ID)
	}

	attachments := make([]gin.H, 0, len(article.Attachments))
	for _, att := range article.Attachments {
		attachments = append(attachments, gin.H{
			"id":         att.ID,
			"name":       att.Name,
			"url":        att.URL,
			"size":       att.Size,
			"createTime": att.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	ctx.JSON(http.StatusOK, utils.Success("获取文章成功", gin.H{
		"id":           article.ID,
		"title":        article.Title,
		"type":         article.Type,
		"categoryIds":  categoryIds,
		"categoryName": strings.Join(categoryNames, "、"),
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
		"publishTime":  formatLocalTime(article.PublishTime),
		"url":          article.URL,
		"columnCount":  len(article.Columns),
		"tagIds":       tagIds,
		"columnIds":    columnIds,
		"attachments":  attachments,
		"createTime":   article.CreatedAt.Format("2006-01-02 15:04:05"),
		"updatedAt":    article.UpdatedAt.Format("2006-01-02 15:04:05"),
	}))
}

func (c *ArticleController) CreateArticle(ctx *gin.Context) {
	var req struct {
		models.Article
		TagIds      []uint `json:"tagIds"`
		CategoryIds []uint `json:"categoryIds"`
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

	err = c.articleService.CreateArticle(&req.Article, req.TagIds, req.CategoryIds)
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
		TagIds      []uint `json:"tagIds"`
		CategoryIds []uint `json:"categoryIds"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: "+err.Error()))
		return
	}

	userID := ctx.GetUint("userID")
	article, err := c.articleService.GetArticleByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章不存在"))
		return
	}
	if strconv.FormatUint(uint64(userID), 10) != article.AuthorCode {
		ctx.JSON(http.StatusOK, utils.Error(1, "无权操作"))
		return
	}
	user, err := c.userService.GetUserByID(userID)
	if err == nil && user != nil {
		req.Article.Author = user.Username
		req.Article.AuthorCode = strconv.FormatUint(uint64(user.ID), 10)
	}

	err = c.articleService.UpdateArticle(uint(id), &req.Article, req.TagIds, req.CategoryIds)
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

func (c *ArticleController) RestartArticleAudit(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章ID无效"))
		return
	}
	userID := ctx.GetUint("userID")
	article, err := c.articleService.GetArticleByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章不存在"))
		return
	}
	if strconv.FormatUint(uint64(userID), 10) != article.AuthorCode {
		ctx.JSON(http.StatusOK, utils.Error(1, "无权操作"))
		return
	}
	err = c.articleService.RestartArticleAudit(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("重新提交审核成功", nil))
}

func (c *ArticleController) WithdrawArticleAudit(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章ID无效"))
		return
	}
	userID := ctx.GetUint("userID")
	article, err := c.articleService.GetArticleByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章不存在"))
		return
	}
	if strconv.FormatUint(uint64(userID), 10) != article.AuthorCode {
		ctx.JSON(http.StatusOK, utils.Error(1, "无权操作"))
		return
	}
	err = c.articleService.WithdrawArticleAudit(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("撤回审核成功", nil))
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
	switch req.AuditStatus {
	case 1:
		// 提交审核
		userID := ctx.GetUint("userID")
		article, err := c.articleService.GetArticleByID(uint(id))
		if err != nil {
			ctx.JSON(http.StatusOK, utils.Error(1, "文章不存在"))
			return
		}
		if strconv.FormatUint(uint64(userID), 10) != article.AuthorCode {
			ctx.JSON(http.StatusOK, utils.Error(1, "无权操作"))
			return
		}
		err = c.articleService.StartArticleAudit(uint(id))
	case 2:
		// 完成审核
		err = c.articleService.CompleteArticleAudit(uint(id))
	default:
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
	article, err := c.articleService.GetArticleByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章不存在"))
		return
	}
	userID := ctx.GetUint("userID")
	type auditProgressItem struct {
		models.ArticleColumnAudit
		CanApprove          bool   `json:"canApprove"`
		CurrentApproverName string `json:"currentApproverName"`
	}
	list := make([]auditProgressItem, 0, len(audits))
	for _, audit := range audits {
		item := auditProgressItem{ArticleColumnAudit: audit}
		if audit.Status == 0 {
			item.CanApprove, _ = c.articleService.CanApproveArticleColumn(uint(id), audit.ColumnID, userID)
			node, err := c.workflowService.GetWorkflowNodeByID(audit.CurrentNodeID)
			if err == nil {
				item.CurrentApproverName = c.articleService.GetApproverName(node, article.AuthorCode)
			}
		}
		list = append(list, item)
	}
	ctx.JSON(http.StatusOK, utils.Success("获取审核进度成功", list))
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

func (c *ArticleController) GetArticleAuditHistory(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章ID无效"))
		return
	}
	columnIDStr := ctx.Query("columnId")
	columnID, err := strconv.ParseUint(columnIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "栏目ID无效"))
		return
	}
	histories, err := c.articleService.GetArticleAuditHistory(uint(id), uint(columnID))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取审核历史失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取审核历史成功", histories))
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
	userID := ctx.GetUint("userID")
	article, err := c.articleService.GetArticleByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章不存在"))
		return
	}
	if strconv.FormatUint(uint64(userID), 10) != article.AuthorCode {
		ctx.JSON(http.StatusOK, utils.Error(1, "无权操作"))
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
	userID := ctx.GetUint("userID")
	article, err := c.articleService.GetArticleByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章不存在"))
		return
	}
	roleIds, _ := c.userService.GetUserRoleIds(userID)
	isAdmin := slices.Contains(roleIds, 1)
	if !isAdmin && strconv.FormatUint(uint64(userID), 10) != article.AuthorCode {
		ctx.JSON(http.StatusOK, utils.Error(1, "无权操作"))
		return
	}
	err = c.articleService.DeleteArticle(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "删除文章失败: "+err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("删除文章成功", nil))
}

func (c *ArticleController) GetMyAuditArticles(ctx *gin.Context) {
	userID := ctx.GetUint("userID")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	articles, total, err := c.articleService.GetMyAuditArticles(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取待审核文章失败"))
		return
	}

	var list []gin.H
	for _, a := range articles {
		list = append(list, gin.H{
			"id":         a.ID,
			"title":      a.Title,
			"author":     a.Author,
			"createTime": a.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	ctx.JSON(http.StatusOK, utils.Success("获取待审核文章成功", gin.H{
		"list":  list,
		"total": total,
	}))
}

func (c *ArticleController) GetArticleColumnPublishes(ctx *gin.Context) {
	articleTitle := ctx.Query("articleTitle")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("pageSize", "10")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize < 1 {
		pageSize = 10
	}

	articles, total, err := c.articleService.GetArticleColumnPublishes(articleTitle, 0, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取静态化状态列表失败"))
		return
	}

	routePath, name, _ := c.articleService.GetDetailPageRoutePath()

	var result []gin.H
	for _, a := range articles {
		result = append(result, gin.H{
			"id":         a.ID,
			"title":      a.Title,
			"routePath":  fmt.Sprintf("%s/%d.html", routePath, a.ID),
			"name":       name,
			"author":     a.Author,
			"source":     a.Source,
			"createTime": a.CreatedAt.Format("2006-01-02 15:04:05"),
			"updatedAt":  a.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	ctx.JSON(http.StatusOK, utils.Success("获取静态化状态列表成功", gin.H{
		"list":     result,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}))
}

func (c *ArticleController) GetArticleAuthorStats(ctx *gin.Context) {
	period := ctx.Query("period")
	if period == "" {
		period = "week"
	}

	stats, total, err := c.articleService.GetArticleAuthorStats(period)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取文章作者统计失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("获取成功", gin.H{
		"list":  stats,
		"total": total,
	}))
}

func formatLocalTime(t *models.LocalTime) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
