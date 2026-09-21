package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/services"
	"server/utils"
)

type LinkController struct {
	linkService *services.LinkService
	userService *services.UserService
}

func NewLinkController() *LinkController {
	return &LinkController{
		linkService: &services.LinkService{},
		userService: &services.UserService{},
	}
}

func (c *LinkController) GetLinks(ctx *gin.Context) {
	name := ctx.Query("name")
	templateIDStr := ctx.Query("templateId")
	columnIDStr := ctx.Query("columnId")
	statusStr := ctx.Query("status")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("pageSize", "10")

	templateID := 0
	if templateIDStr != "" {
		if id, err := strconv.Atoi(templateIDStr); err == nil {
			templateID = id
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
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize < 1 {
		pageSize = 10
	}

	links, total, err := c.linkService.GetLinks(name, templateID, columnID, status, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取友链列表失败"))
		return
	}

	// 补充模板和栏目名称
	var templateIDs []uint
	var columnIDs []uint
	for _, l := range links {
		if l.TemplateID > 0 {
			templateIDs = append(templateIDs, l.TemplateID)
		}
		if l.ColumnID > 0 {
			columnIDs = append(columnIDs, l.ColumnID)
		}
	}

	templateMap := make(map[uint]string)
	columnMap := make(map[uint]string)
	if len(templateIDs) > 0 {
		var templates []models.Template
		utils.DB.Where("id IN ?", templateIDs).Find(&templates)
		for _, t := range templates {
			templateMap[t.ID] = t.Name
		}
	}
	if len(columnIDs) > 0 {
		var columns []models.Column
		utils.DB.Where("id IN ?", columnIDs).Find(&columns)
		for _, col := range columns {
			columnMap[col.ID] = col.Name
		}
	}

	var list []gin.H
	for _, l := range links {
		list = append(list, gin.H{
			"id":           l.ID,
			"name":         l.Name,
			"url":          l.Url,
			"logo":         l.Logo,
			"description":  l.Description,
			"templateId":   l.TemplateID,
			"templateName": templateMap[l.TemplateID],
			"columnId":     l.ColumnID,
			"columnName":   columnMap[l.ColumnID],
			"sort":         l.Sort,
			"status":       l.Status,
			"author":       l.Author,
			"authorCode":   l.AuthorCode,
			"createdAt":    l.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	ctx.JSON(http.StatusOK, utils.Success("获取友链列表成功", utils.PageData(list, total, page, pageSize)))
}

func (c *LinkController) CreateLink(ctx *gin.Context) {
	var req models.Link
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}
	if err := services.ValidateTemplateColumn(req.TemplateID, req.ColumnID); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SafeErrText(err)))
		return
	}
	userID := ctx.GetUint("userID")
	user, err := c.userService.GetUserByID(userID)
	if err == nil && user != nil {
		req.Author = user.Username
		req.AuthorCode = strconv.FormatUint(uint64(user.ID), 10)
	}
	err = c.linkService.CreateLink(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("创建友链失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("创建友链成功", nil))
}

func (c *LinkController) UpdateLink(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "友链ID无效"))
		return
	}
	var req models.Link
	if err = ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}
	if err := services.ValidateTemplateColumn(req.TemplateID, req.ColumnID); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SafeErrText(err)))
		return
	}
	userID := ctx.GetUint("userID")
	user, err := c.userService.GetUserByID(userID)
	if err == nil && user != nil {
		req.Author = user.Username
		req.AuthorCode = strconv.FormatUint(uint64(user.ID), 10)
	}
	err = c.linkService.UpdateLink(uint(id), &req)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("更新友链失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新友链成功", nil))
}

func (c *LinkController) UpdateLinkStatus(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "友链ID无效"))
		return
	}
	var req struct {
		Status int `json:"status"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误"))
		return
	}
	err = c.linkService.UpdateLinkStatus(uint(id), req.Status)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "更新状态失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新状态成功", nil))
}

func (c *LinkController) DeleteLink(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "友链ID无效"))
		return
	}
	err = c.linkService.DeleteLink(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("删除友链失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("删除友链成功", nil))
}
