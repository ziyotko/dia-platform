package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/services"
	"server/utils"
)

type ArticleController struct {
	articleService      *services.ArticleService
	userService         *services.UserService
	workflowService     *services.WorkflowService
	staticJobController *StaticJobController
}

func NewArticleController() *ArticleController {
	return &ArticleController{
		articleService:      &services.ArticleService{},
		userService:         &services.UserService{},
		workflowService:     &services.WorkflowService{},
		staticJobController: NewStaticJobController(),
	}
}

func (c *ArticleController) GetArticles(ctx *gin.Context) {
	title := ctx.Query("title")
	categoryIDStr := ctx.Query("categoryId")
	tagIDStr := ctx.Query("tagId")
	columnIDStr := ctx.Query("columnId")
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
	columnID := 0
	if columnIDStr != "" {
		if id, err := strconv.Atoi(columnIDStr); err == nil {
			columnID = id
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

	// 非管理员只能看到自己的文章（“内容作者”角色仅限自有内容；审核人通过「待审核」页/详情接口查看待审文章）
	authorCodeScope := ""
	currentUserID := ctx.GetUint("userID")
	if !models.HasAdminRoleIDs(c.userService.MustGetUserRoleIds(currentUserID)) {
		authorCodeScope = strconv.FormatUint(uint64(currentUserID), 10)
	}

	articles, total, err := c.articleService.GetArticles(title, categoryID, tagID, columnID, status, auditStatus, articleType, author, authorCodeScope, source, page, pageSize)
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
				"id":        att.ID,
				"name":      att.Name,
				"url":       att.URL,
				"size":      att.Size,
				"createdAt": att.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
		list = append(list, gin.H{
			"id":           a.ID,
			"title":        a.Title,
			"type":         a.Type,
			"summary":      a.Summary,
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
			"createdAt":    a.CreatedAt.Format("2006-01-02 15:04:05"),
			"updatedAt":    a.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	ctx.JSON(http.StatusOK, utils.Success("获取文章列表成功", utils.PageData(list, total, page, pageSize)))
}

// PublicSearchArticles 开放搜索（无需认证）：按查询条件搜索已发布文章，返回 article 表数据但不含 content 字段
func (c *ArticleController) PublicSearchArticles(ctx *gin.Context) {
	title := ctx.Query("title")
	categoryIDStr := ctx.Query("categoryId")
	tagIDStr := ctx.Query("tagId")
	typeStr := ctx.Query("type")
	author := ctx.Query("author")
	source := ctx.Query("source")

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
	articleType := 0
	if typeStr != "" {
		if t, err := strconv.Atoi(typeStr); err == nil {
			articleType = t
		}
	}
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if pageSize < 1 {
		pageSize = 10
	}

	articles, total, err := c.articleService.SearchPublishedArticles(title, categoryID, tagID, articleType, author, source, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "搜索文章失败"))
		return
	}

	var list []gin.H
	for _, a := range articles {
		categoryIds := make([]uint, 0, len(a.Categories))
		for _, c := range a.Categories {
			categoryIds = append(categoryIds, c.ID)
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
			"type":         a.Type,
			"summary":      a.Summary,
			"status":       a.Status,
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
			"categoryIds":  categoryIds,
			"tagIds":       tagIds,
			"columnIds":    columnIds,
			"createdAt":    a.CreatedAt.Format("2006-01-02 15:04:05"),
			"updatedAt":    a.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	ctx.JSON(http.StatusOK, utils.Success("搜索成功", utils.PageData(list, total, page, pageSize)))
}

// canViewArticle 文章可见性判定（统一口径）：管理员、作者本人，或该文章存在「待我审批」的栏目。
// 正文（GetArticleByID）与审核进度/审核历史（含审批人姓名、审批/驳回意见）必须使用同一判定，
// 否则任何拿到 /articles 菜单的用户只需遍历 ID 即可读到他人文章的审批信息。
func (c *ArticleController) canViewArticle(article *models.Article, userID uint) bool {
	if models.HasAdminRoleIDs(c.userService.MustGetUserRoleIds(userID)) {
		return true
	}
	if strconv.FormatUint(uint64(userID), 10) == article.AuthorCode {
		return true
	}
	allowed, err := c.articleService.CanApproveAnyColumn(article.ID, userID)
	return err == nil && allowed
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

	// 可见性：管理员/作者本人/该文章存在待我审批的栏目时可以查看正文
	userID := ctx.GetUint("userID")
	if !c.canViewArticle(article, userID) {
		ctx.JSON(http.StatusOK, utils.Error(1, "无权查看该文章"))
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
			"id":        att.ID,
			"name":      att.Name,
			"url":       att.URL,
			"size":      att.Size,
			"createdAt": att.CreatedAt.Format("2006-01-02 15:04:05"),
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
		"createdAt":    article.CreatedAt.Format("2006-01-02 15:04:05"),
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
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
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
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("创建文章失败", err)))
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
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
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
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("更新文章失败", err)))
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

	// 仅允许草稿(0)/下线(2)；发布(1) 必须走审核流程（CompleteArticleAudit），防止绕过审核直接发布
	if req.Status != 0 && req.Status != 2 {
		ctx.JSON(http.StatusOK, utils.Error(1, "不支持的状态变更，文章发布请通过审核流程"))
		return
	}

	// 权限校验：仅作者本人或管理员可以修改文章发布状态
	userID := ctx.GetUint("userID")
	article, err := c.articleService.GetArticleByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章不存在"))
		return
	}
	roleIds, _ := c.userService.GetUserRoleIds(userID)
	isAdmin := models.HasAdminRoleIDs(roleIds)
	if !isAdmin && strconv.FormatUint(uint64(userID), 10) != article.AuthorCode {
		ctx.JSON(http.StatusOK, utils.Error(1, "无权操作"))
		return
	}

	err = c.articleService.UpdateArticleStatus(uint(id), req.Status)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "更新文章状态失败"))
		return
	}

	// 文章下线（status == 2）时，调用静态化功能删除该文章的静态文件
	// 静态化程序接口：DELETE /api/static/article?id={文章ID}&path={静态化输出路径}
	// 代理处理见 static_job_controller.go DeleteArticleStatic
	if req.Status == 2 {
		c.staticJobController.DeleteArticleStaticByID(ctx, idStr)
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
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SafeErrText(err)))
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
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SafeErrText(err)))
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
	userID := ctx.GetUint("userID")
	switch req.AuditStatus {
	case 1:
		// 提交审核：仅作者本人
		var article *models.Article
		article, err = c.articleService.GetArticleByID(uint(id))
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
		// 完成审核：仅管理员（服务层会校验所有栏目审核均已通过）
		roleIds, _ := c.userService.GetUserRoleIds(userID)
		if !models.HasAdminRoleIDs(roleIds) {
			ctx.JSON(http.StatusOK, utils.Error(1, "无权限执行该操作"))
			return
		}
		err = c.articleService.CompleteArticleAudit(uint(id))
	default:
		// 审核状态只能通过「提交审核(1)」与「完成审核(2)」两个动作，或撤回/重新提交专用接口变更，
		// 不再允许任意设置 audit_status（旧实现可写入任意值且不维护审核记录）。
		ctx.JSON(http.StatusOK, utils.Error(1, "不支持的审核操作，auditStatus 仅支持 1=提交审核、2=完成审核"))
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("审核文章失败", err)))
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
	if !c.canViewArticle(article, userID) {
		ctx.JSON(http.StatusOK, utils.Error(1, "无权查看该文章"))
		return
	}
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
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SafeErrText(err)))
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
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SafeErrText(err)))
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
	// 审核历史包含审批人姓名与审批/驳回意见，须先通过文章可见性校验（与正文接口同口径）
	article, err := c.articleService.GetArticleByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章不存在"))
		return
	}
	if !c.canViewArticle(article, ctx.GetUint("userID")) {
		ctx.JSON(http.StatusOK, utils.Error(1, "无权查看该文章"))
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
	isAdmin := models.HasAdminRoleIDs(roleIds)
	if !isAdmin && strconv.FormatUint(uint64(userID), 10) != article.AuthorCode {
		ctx.JSON(http.StatusOK, utils.Error(1, "无权操作"))
		return
	}
	err = c.articleService.DeleteArticle(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("删除文章失败", err)))
		return
	}

	// 删除文章后，调用静态化功能删除该文章的静态文件
	// 静态化程序接口：DELETE /api/static/article?id={文章ID}&path={静态化输出路径}
	// 代理处理见 static_job_controller.go DeleteArticleStatic
	c.staticJobController.DeleteArticleStaticByID(ctx, idStr)

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
			"id":        a.ID,
			"title":     a.Title,
			"author":    a.Author,
			"createdAt": a.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	ctx.JSON(http.StatusOK, utils.Success("获取待审核文章成功", utils.PageData(list, total, page, pageSize)))
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
	// 栏目筛选：原先恒传 0，导致服务层与前端 API 声明的 columnId 筛选静默失效
	columnID, _ := strconv.ParseUint(ctx.Query("columnId"), 10, 32)

	articles, total, err := c.articleService.GetArticleColumnPublishes(articleTitle, uint(columnID), page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取静态化状态列表失败"))
		return
	}

	routePath, name, _ := c.articleService.GetDetailPageRoutePath()

	var result []gin.H
	for _, a := range articles {
		result = append(result, gin.H{
			"id":        a.ID,
			"title":     a.Title,
			"routePath": fmt.Sprintf("%s/%d.html", routePath, a.ID),
			"name":      name,
			"author":    a.Author,
			"source":    a.Source,
			"createdAt": a.CreatedAt.Format("2006-01-02 15:04:05"),
			"updatedAt": a.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	ctx.JSON(http.StatusOK, utils.Success("获取静态化状态列表成功", utils.PageData(result, total, page, pageSize)))
}

func (c *ArticleController) GetArticleAuthorStats(ctx *gin.Context) {
	period := ctx.Query("period")
	if period == "" {
		period = "week"
	}

	// 归属过滤：与文章列表同一口径，非管理员只统计自己的文章（否则可读出全站作者发文量）
	userID := ctx.GetUint("userID")
	authorCode := ""
	if !models.HasAdminRoleIDs(c.userService.MustGetUserRoleIds(userID)) {
		authorCode = strconv.FormatUint(uint64(userID), 10)
	}

	stats, total, err := c.articleService.GetArticleAuthorStats(period, authorCode)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取文章作者统计失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("获取成功", utils.AllData(stats, total)))
}

func formatLocalTime(t *models.LocalTime) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
