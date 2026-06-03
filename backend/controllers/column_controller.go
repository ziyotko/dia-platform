package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/services"
	"server/utils"
)

type ColumnController struct {
	columnService *services.ColumnService
}

func NewColumnController() *ColumnController {
	return &ColumnController{
		columnService: &services.ColumnService{},
	}
}

func (c *ColumnController) GetColumns(ctx *gin.Context) {
	pageIDStr := ctx.Query("pageId")
	parentIDStr := ctx.Query("parentId")
	var pageID uint
	var parentID uint
	if pageIDStr != "" {
		if id, err := strconv.ParseUint(pageIDStr, 10, 32); err == nil {
			pageID = uint(id)
		}
	}
	if parentIDStr != "" {
		if id, err := strconv.ParseUint(parentIDStr, 10, 32); err == nil {
			parentID = uint(id)
		}
	}
	columns, err := c.columnService.GetColumns(pageID, parentID)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取栏目列表失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取栏目列表成功", columns))
}

func (c *ColumnController) CreateColumn(ctx *gin.Context) {
	var req models.Column
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: "+err.Error()))
		return
	}
	err := c.columnService.CreateColumn(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "创建栏目失败: "+err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("创建栏目成功", nil))
}

func (c *ColumnController) UpdateColumn(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "栏目ID无效"))
		return
	}
	var req models.Column
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: "+err.Error()))
		return
	}
	err = c.columnService.UpdateColumn(uint(id), &req)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "更新栏目失败: "+err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新栏目成功", nil))
}

func (c *ColumnController) DeleteColumn(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "栏目ID无效"))
		return
	}
	err = c.columnService.DeleteColumn(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "删除栏目失败: "+err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("删除栏目成功", nil))
}
