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

type TemplateController struct {
	templateService *services.TemplateService
}

type TemplateListItem struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Type        string `json:"type"`
	RoutePath   string `json:"routePath"`
	Description string `json:"description"`
	Status      int    `json:"status"`
	SourceCode  string `json:"sourceCode"`
	Layout      string `json:"layout"`
	ColumnCount int    `json:"columnCount"`
	ColumnName  string `json:"columnName"`
	CreatedAt   string `json:"createdAt"`
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

	// all=1 时不分页返回全量（供栏目绑定模板等下拉使用）
	all := ctx.Query("all") == "1" || strings.EqualFold(ctx.Query("all"), "true")
	if all {
		page, pageSize = 1, 0
	} else {
		if page < 1 {
			page = 1
		}
		if pageSize < 1 {
			pageSize = 10
		}
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
			RoutePath:   t.RoutePath,
			Description: t.Description,
			Status:      t.Status,
			SourceCode:  t.SourceCode,
			Layout:      t.Layout,
			ColumnCount: result.ColumnCounts[t.ID],
			ColumnName:  result.ColumnNames[t.ID],
			CreatedAt:   t.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	ctx.JSON(http.StatusOK, utils.Success("获取模板列表成功", utils.PageData(list, result.Total, page, pageSize)))
}

func (c *TemplateController) CreateTemplate(ctx *gin.Context) {
	var req models.Template
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}

	err := c.templateService.CreateTemplate(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("创建模板失败", err)))
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
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}

	// 前端编辑弹窗会提交「模板编码 / 访问路径」（模板即页面，code/route_path 参与静态化路径与预览地址拼接），
	// 原先漏写这两列：改「访问路径」返回成功但库不变，详情页静态文件路径与 /articles/column-publishes
	// 的 routePath/url 会一直沿用旧值，且界面上无法修正。
	updates := map[string]any{
		"name":        req.Name,
		"code":        req.Code,
		"type":        req.Type,
		"route_path":  req.RoutePath,
		"description": req.Description,
		"status":      req.Status,
	}

	err = c.templateService.UpdateTemplate(uint(id), updates)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("更新模板失败", err)))
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
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("删除模板失败", err)))
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
	if err = ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误"))
		return
	}
	if req.Status != 0 && req.Status != 1 {
		ctx.JSON(http.StatusOK, utils.Error(1, "状态值无效"))
		return
	}

	err = c.templateService.UpdateTemplateStatus(uint(id), req.Status)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("更新状态失败", err)))
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

	// 字段用指针接收：仅在请求体显式携带该字段时才更新。
	// 原先 layout 为 string，前端每次「保存设计」都发送 layout:""，会把模板已有的 layout 清空
	// （layout 是模板预览的兜底数据源，清空后预览只剩空态）。
	var req struct {
		SourceCode *string `json:"sourceCode"`
		Layout     *string `json:"layout"`
	}
	if err = ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误"))
		return
	}

	updates := map[string]any{}
	if req.SourceCode != nil {
		updates["source_code"] = *req.SourceCode
	}
	if req.Layout != nil {
		updates["layout"] = *req.Layout
	}
	if len(updates) == 0 {
		ctx.JSON(http.StatusOK, utils.Error(1, "没有需要保存的内容"))
		return
	}

	err = c.templateService.UpdateTemplate(uint(id), updates)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("保存设计失败", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("保存设计成功", nil))
}
