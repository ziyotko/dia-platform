package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/services"
	"server/utils"
)

type TemplateController struct {
	templateService *services.TemplateService
}

type TemplateListItem struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Status      int    `json:"status"`
	SourceCode  string `json:"sourceCode"`
	Layout      string `json:"layout"`
	PageCount   int    `json:"pageCount"`
	CreateTime  string `json:"createTime"`
}

func NewTemplateController() *TemplateController {
	return &TemplateController{
		templateService: &services.TemplateService{},
	}
}

func (c *TemplateController) GetTemplates(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	name := ctx.Query("name")
	ttype := ctx.Query("type")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	result, err := c.templateService.GetTemplateList(page, pageSize, name, ttype)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取模板列表失败"))
		return
	}

	list := make([]TemplateListItem, 0, len(result.List))
	for _, t := range result.List {
		list = append(list, TemplateListItem{
			ID:          t.ID,
			Name:        t.Name,
			Code:        t.Code,
			Type:        t.Type,
			Description: t.Description,
			Status:      t.Status,
			SourceCode:  t.SourceCode,
			Layout:      t.Layout,
			PageCount:   t.PageCount,
			CreateTime:  t.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	ctx.JSON(http.StatusOK, utils.Success("获取模板列表成功", gin.H{
		"list":  list,
		"total": result.Total,
	}))
}

func (c *TemplateController) CreateTemplate(ctx *gin.Context) {
	var req models.Template
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: "+err.Error()))
		return
	}

	err := c.templateService.CreateTemplate(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "创建模板失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("创建模板成功", nil))
}

func (c *TemplateController) UpdateTemplate(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "模板ID无效"))
		return
	}

	var req models.Template
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: "+err.Error()))
		return
	}

	updates := map[string]interface{}{
		"name":        req.Name,
		"code":        req.Code,
		"type":        req.Type,
		"description": req.Description,
		"status":      req.Status,
	}

	err = c.templateService.UpdateTemplate(uint(id), updates)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "更新模板失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("更新模板成功", nil))
}

func (c *TemplateController) DeleteTemplate(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "模板ID无效"))
		return
	}

	err = c.templateService.DeleteTemplate(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "删除模板失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("删除模板成功", nil))
}

func (c *TemplateController) UpdateTemplateStatus(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "模板ID无效"))
		return
	}

	var req struct {
		Status int `json:"status"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误"))
		return
	}

	err = c.templateService.UpdateTemplate(uint(id), map[string]interface{}{"status": req.Status})
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "更新状态失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("更新状态成功", nil))
}

func (c *TemplateController) SaveTemplateDesign(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "模板ID无效"))
		return
	}

	var req struct {
		SourceCode string `json:"sourceCode"`
		Layout     string `json:"layout"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误"))
		return
	}

	err = c.templateService.UpdateTemplate(uint(id), map[string]interface{}{
		"source_code": req.SourceCode,
		"layout":      req.Layout,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "保存设计失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("保存设计成功", nil))
}
