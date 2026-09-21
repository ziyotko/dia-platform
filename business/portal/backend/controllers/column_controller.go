package controllers

import (
	"net/http"
	"strconv"
	"strings"

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
	templateIDStr := ctx.Query("templateId")
	parentIDStr := ctx.Query("parentId")
	displayTypeStr := ctx.Query("displayType")
	var templateID uint
	var parentID uint
	var displayType int
	if templateIDStr != "" {
		if id, err := strconv.ParseUint(templateIDStr, 10, 32); err == nil {
			templateID = uint(id)
		}
	}
	if parentIDStr != "" {
		if id, err := strconv.ParseUint(parentIDStr, 10, 32); err == nil {
			parentID = uint(id)
		}
	}
	if displayTypeStr != "" {
		if dt, err := strconv.Atoi(displayTypeStr); err == nil {
			displayType = dt
		}
	}
	columns, err := c.columnService.GetColumns(templateID, parentID, displayType)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取栏目列表失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取栏目列表成功", columns))
}

// GetColumnPublishes 获取栏目发布（静态化）列表
func (c *ColumnController) GetColumnPublishes(ctx *gin.Context) {
	columns, err := c.columnService.GetColumnPublishes()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取栏目发布列表失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取栏目发布列表成功", utils.AllData(columns, int64(len(columns)))))
}

func (c *ColumnController) CreateColumn(ctx *gin.Context) {
	var req models.Column
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}
	err := c.columnService.CreateColumn(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("创建栏目失败", err)))
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
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}
	err = c.columnService.UpdateColumn(uint(id), &req)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("更新栏目失败", err)))
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
		if strings.Contains(err.Error(), "Cannot delete or update a parent row") {
			ctx.JSON(http.StatusOK, utils.Error(1, "该栏目存在关联数据（子栏目或已发布文章），无法直接删除，请先解除关联"))
			return
		}
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("删除栏目失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("删除栏目成功", nil))
}
