package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/services"
	"server/utils"
)

type CategoryController struct {
	categoryService *services.CategoryService
}

func NewCategoryController() *CategoryController {
	return &CategoryController{
		categoryService: &services.CategoryService{},
	}
}

func (c *CategoryController) GetCategories(ctx *gin.Context) {
	name := ctx.Query("name")
	statusStr := ctx.Query("status")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("pageSize", "10")

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

	categories, total, err := c.categoryService.GetCategories(name, status, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取分类列表失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取分类列表成功", gin.H{
		"list":     categories,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}))
}

func (c *CategoryController) GetAllCategories(ctx *gin.Context) {
	categories, err := c.categoryService.GetAllCategories()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取分类列表失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取分类列表成功", categories))
}

func (c *CategoryController) CreateCategory(ctx *gin.Context) {
	var req models.Category
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: "+err.Error()))
		return
	}
	err := c.categoryService.CreateCategory(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "创建分类失败: "+err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("创建分类成功", nil))
}

func (c *CategoryController) UpdateCategory(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "分类ID无效"))
		return
	}
	var req models.Category
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: "+err.Error()))
		return
	}
	err = c.categoryService.UpdateCategory(uint(id), &req)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "更新分类失败: "+err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新分类成功", nil))
}

func (c *CategoryController) UpdateCategoryStatus(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "分类ID无效"))
		return
	}
	var req struct {
		Status int `json:"status"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误"))
		return
	}
	err = c.categoryService.UpdateCategoryStatus(uint(id), req.Status)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "更新分类状态失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新分类状态成功", nil))
}

func (c *CategoryController) DeleteCategory(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "分类ID无效"))
		return
	}
	err = c.categoryService.DeleteCategory(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "删除分类失败: "+err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("删除分类成功", nil))
}
