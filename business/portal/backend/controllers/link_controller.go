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
	pageIDStr := ctx.Query("pageId")
	columnIDStr := ctx.Query("columnId")
	statusStr := ctx.Query("status")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("pageSize", "10")

	pageID := 0
	if pageIDStr != "" {
		if id, err := strconv.Atoi(pageIDStr); err == nil {
			pageID = id
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

	links, total, err := c.linkService.GetLinks(name, pageID, columnID, status, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取友链列表失败"))
		return
	}

	// 补充页面和栏目名称
	var pageIDs []uint
	var columnIDs []uint
	for _, l := range links {
		if l.PageID > 0 {
			pageIDs = append(pageIDs, l.PageID)
		}
		if l.ColumnID > 0 {
			columnIDs = append(columnIDs, l.ColumnID)
		}
	}

	pageMap := make(map[uint]string)
	columnMap := make(map[uint]string)
	if len(pageIDs) > 0 {
		var pages []models.Page
		utils.DB.Where("id IN ?", pageIDs).Find(&pages)
		for _, p := range pages {
			pageMap[p.ID] = p.Name
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
			"id":          l.ID,
			"name":        l.Name,
			"url":         l.Url,
			"logo":        l.Logo,
			"description": l.Description,
			"pageId":      l.PageID,
			"pageName":    pageMap[l.PageID],
			"columnId":    l.ColumnID,
			"columnName":  columnMap[l.ColumnID],
			"sort":        l.Sort,
			"status":      l.Status,
			"author":      l.Author,
			"authorCode":  l.AuthorCode,
			"createdAt":   l.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	ctx.JSON(http.StatusOK, utils.Success("获取友链列表成功", utils.PageData(list, total, page, pageSize)))
}

func (c *LinkController) GetLinkByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "友链ID无效"))
		return
	}
	link, err := c.linkService.GetLinkByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取友链失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("获取友链成功", gin.H{
		"id":          link.ID,
		"name":        link.Name,
		"url":         link.Url,
		"logo":        link.Logo,
		"description": link.Description,
		"pageId":      link.PageID,
		"columnId":    link.ColumnID,
		"sort":        link.Sort,
		"status":      link.Status,
		"author":      link.Author,
		"authorCode":  link.AuthorCode,
	}))
}

func (c *LinkController) CreateLink(ctx *gin.Context) {
	var req models.Link
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
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
