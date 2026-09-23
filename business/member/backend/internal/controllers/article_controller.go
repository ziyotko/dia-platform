package controllers

import (
	"member/internal/middleware"
	"member/internal/service"
	"member/pkg/response"

	"github.com/gin-gonic/gin"
)

type ArticleController struct {
	articleService service.ArticleService
	catService     service.ArticleCategoryService
}

// ---- Article Category ----

func (ctrl *ArticleController) ListCategories(c *gin.Context) {
	cats, err := ctrl.catService.ListCategories()
	if err != nil {
		response.ServerErrorFrom(c, err)
		return
	}
	response.Success(c, cats)
}

func (ctrl *ArticleController) CreateCategory(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
		Sort int    `json:"sort"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	cat, err := ctrl.catService.CreateCategory(req.Name, req.Sort)
	if err != nil {
		response.BadRequestFrom(c, err)
		return
	}
	response.SuccessWithMessage(c, "创建成功", cat)
}

func (ctrl *ArticleController) UpdateCategory(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req struct {
		Name string `json:"name" binding:"required"`
		Sort int    `json:"sort"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.catService.UpdateCategory(id, req.Name, req.Sort); err != nil {
		response.BadRequestFrom(c, err)
		return
	}
	response.SuccessWithMessage(c, "更新成功", nil)
}

func (ctrl *ArticleController) DeleteCategory(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.catService.DeleteCategory(id); err != nil {
		response.BadRequestFrom(c, err)
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}

// ---- Articles (Member) ----

func (ctrl *ArticleController) CreateArticle(c *gin.Context) {
	memberID := middleware.GetMemberID(c)
	var req service.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整信息")
		return
	}
	article, err := ctrl.articleService.CreateArticle(memberID, req)
	if err != nil {
		response.BadRequestFrom(c, err)
		return
	}
	response.SuccessWithMessage(c, "创建成功", article)
}

func (ctrl *ArticleController) UpdateArticle(c *gin.Context) {
	memberID := middleware.GetMemberID(c)
	articleID := parseUint(c.Param("id"))
	var req service.UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.articleService.UpdateArticle(memberID, articleID, req); err != nil {
		response.BadRequestFrom(c, err)
		return
	}
	response.SuccessWithMessage(c, "更新成功", nil)
}

func (ctrl *ArticleController) DeleteArticle(c *gin.Context) {
	memberID := middleware.GetMemberID(c)
	articleID := parseUint(c.Param("id"))
	if err := ctrl.articleService.DeleteArticle(memberID, articleID); err != nil {
		response.BadRequestFrom(c, err)
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}

func (ctrl *ArticleController) GetArticle(c *gin.Context) {
	memberID := middleware.GetMemberID(c)
	id := parseUint(c.Param("id"))
	article, err := ctrl.articleService.GetArticle(id, memberID)
	if err != nil {
		response.NotFoundFrom(c, err)
		return
	}
	response.Success(c, article)
}

func (ctrl *ArticleController) GetMyArticles(c *gin.Context) {
	memberID := middleware.GetMemberID(c)
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 10)
	status := c.Query("status")
	categoryIDStr := c.Query("category_id")
	var categoryID uint64
	if categoryIDStr != "" {
		categoryID = parseUint(categoryIDStr)
	}

	articles, total, err := ctrl.articleService.GetMyArticles(memberID, page, size, status, categoryID)
	if err != nil {
		response.ServerErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"list":  articles,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// ---- Articles (Public) ----

func (ctrl *ArticleController) ListArticles(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 10)
	categoryIDStr := c.Query("category_id")
	keyword := c.Query("keyword")
	var categoryID uint64
	if categoryIDStr != "" {
		categoryID = parseUint(categoryIDStr)
	}

	articles, total, err := ctrl.articleService.ListArticles(page, size, categoryID, keyword)
	if err != nil {
		response.ServerErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"list":  articles,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

func (ctrl *ArticleController) GetPublishedArticle(c *gin.Context) {
	id := parseUint(c.Param("id"))
	article, err := ctrl.articleService.GetPublishedArticle(id)
	if err != nil {
		response.NotFoundFrom(c, err)
		return
	}
	response.Success(c, article)
}

// ---- Articles (Admin) ----

func (ctrl *ArticleController) ReviewArticle(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req struct {
		Approved bool   `json:"approved"`
		Comment  string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.articleService.ReviewArticle(id, req.Approved, req.Comment); err != nil {
		response.BadRequestFrom(c, err)
		return
	}
	response.SuccessWithMessage(c, "审核完成", nil)
}

func (ctrl *ArticleController) ListAllArticles(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 10)
	status := c.Query("status")

	articles, total, err := ctrl.articleService.ListAllArticles(page, size, status)
	if err != nil {
		response.ServerErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"list":  articles,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

func (ctrl *ArticleController) AdminDeleteArticle(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.articleService.AdminDeleteArticle(id); err != nil {
		response.BadRequestFrom(c, err)
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}
