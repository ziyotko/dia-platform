package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"portal/internal/models"
	"portal/internal/services"
	"portal/pkg/utils"
)

// MemberZoneController 会员专区：会员栏目（分类）维护 + 会员专属内容（新闻/数据/视频）发布。
// 会员栏目为管理员维护；内容由作者本人维护（管理员可代管全部）。
type MemberZoneController struct {
	columnService  *services.MemberColumnService
	contentService *services.MemberContentService
	userService    *services.UserService
}

func NewMemberZoneController() *MemberZoneController {
	return &MemberZoneController{
		columnService:  &services.MemberColumnService{},
		contentService: &services.MemberContentService{},
		userService:    &services.UserService{},
	}
}

// memberZoneQueryInt 解析整数查询参数，非法/为空时回退默认值（-1 常用于表示「全部」）。
func memberZoneQueryInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	if v, err := strconv.Atoi(raw); err == nil {
		return v
	}
	return fallback
}

// memberZoneQueryPage 解析分页参数并做下界保护。
func memberZoneQueryPage(raw string, fallback int) int {
	v := memberZoneQueryInt(raw, fallback)
	if v < 1 {
		return fallback
	}
	return v
}

// ============================ 会员栏目（会员栏目分类） ============================

func (c *MemberZoneController) GetMemberColumns(ctx *gin.Context) {
	name := ctx.Query("name")
	status := memberZoneQueryInt(ctx.Query("status"), -1)
	page := memberZoneQueryPage(ctx.DefaultQuery("page", "1"), 1)
	pageSize := memberZoneQueryPage(ctx.DefaultQuery("pageSize", "10"), 10)

	columns, total, err := c.columnService.GetMemberColumns(name, status, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取会员栏目列表失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取会员栏目列表成功", utils.PageData(columns, total, page, pageSize)))
}

func (c *MemberZoneController) CreateMemberColumn(ctx *gin.Context) {
	var req models.MemberColumn
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}
	if err := c.columnService.CreateMemberColumn(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("创建会员栏目失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("创建会员栏目成功", nil))
}

func (c *MemberZoneController) UpdateMemberColumn(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "会员栏目ID无效"))
		return
	}
	var req models.MemberColumn
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}
	if err := c.columnService.UpdateMemberColumn(uint(id), &req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("更新会员栏目失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新会员栏目成功", nil))
}

func (c *MemberZoneController) UpdateMemberColumnStatus(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "会员栏目ID无效"))
		return
	}
	var req struct {
		Status int `json:"status"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误"))
		return
	}
	if req.Status != 0 && req.Status != 1 {
		ctx.JSON(http.StatusOK, utils.Error(1, "状态值无效"))
		return
	}
	if err := c.columnService.UpdateMemberColumnStatus(uint(id), req.Status); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "更新会员栏目状态失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新会员栏目状态成功", nil))
}

func (c *MemberZoneController) DeleteMemberColumn(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "会员栏目ID无效"))
		return
	}
	if err := c.columnService.DeleteMemberColumn(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("删除会员栏目失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("删除会员栏目成功", nil))
}

// ============================ 会员专属内容 ============================

// canAccessMemberContent 管理员或内容作者本人可读写该内容。
func (c *MemberZoneController) canAccessMemberContent(content *models.MemberContent, userID uint) bool {
	if models.HasAdminRoleIDs(c.userService.MustGetUserRoleIds(userID)) {
		return true
	}
	return strconv.FormatUint(uint64(userID), 10) == content.AuthorCode
}

func (c *MemberZoneController) GetMemberContents(ctx *gin.Context) {
	title := ctx.Query("title")
	columnID := memberZoneQueryInt(ctx.Query("columnId"), 0)
	contentType := memberZoneQueryInt(ctx.Query("type"), 0)
	status := memberZoneQueryInt(ctx.Query("status"), -1)
	page := memberZoneQueryPage(ctx.DefaultQuery("page", "1"), 1)
	pageSize := memberZoneQueryPage(ctx.DefaultQuery("pageSize", "10"), 10)

	userID := ctx.GetUint("userID")
	authorCodeScope := ""
	if !models.HasAdminRoleIDs(c.userService.MustGetUserRoleIds(userID)) {
		authorCodeScope = strconv.FormatUint(uint64(userID), 10)
	}

	contents, total, err := c.contentService.GetMemberContents(title, columnID, contentType, status, authorCodeScope, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取会员内容列表失败"))
		return
	}

	ids := make([]uint, 0, len(contents))
	for _, item := range contents {
		ids = append(ids, item.MemberColumnID)
	}
	columnNames, err := c.contentService.GetMemberColumnNames(ids)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取会员内容列表失败"))
		return
	}

	list := make([]gin.H, 0, len(contents))
	for _, item := range contents {
		list = append(list, gin.H{
			"id":               item.ID,
			"memberColumnId":   item.MemberColumnID,
			"memberColumnName": columnNames[item.MemberColumnID],
			"title":            item.Title,
			"type":             item.Type,
			"summary":          item.Summary,
			"cover":            item.Cover,
			"videoUrl":         item.VideoURL,
			"attachmentName":   item.AttachmentName,
			"attachmentUrl":    item.AttachmentURL,
			"author":           item.Author,
			"authorCode":       item.AuthorCode,
			"source":           item.Source,
			"publishTime":      formatLocalTime(item.PublishTime),
			"status":           item.Status,
			"isTop":            item.IsTop,
			"viewCount":        item.ViewCount,
			"createdAt":        item.CreatedAt.Format("2006-01-02 15:04:05"),
			"updatedAt":        item.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	ctx.JSON(http.StatusOK, utils.Success("获取会员内容列表成功", utils.PageData(list, total, page, pageSize)))
}

func (c *MemberZoneController) GetMemberContentByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "内容ID无效"))
		return
	}
	content, err := c.contentService.GetMemberContentByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取会员内容失败"))
		return
	}
	userID := ctx.GetUint("userID")
	if !c.canAccessMemberContent(content, userID) {
		ctx.JSON(http.StatusOK, utils.Error(1, "无权查看该内容"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取会员内容成功", gin.H{
		"id":             content.ID,
		"memberColumnId": content.MemberColumnID,
		"title":          content.Title,
		"type":           content.Type,
		"summary":        content.Summary,
		"content":        content.Content,
		"cover":          content.Cover,
		"videoUrl":       content.VideoURL,
		"attachmentName": content.AttachmentName,
		"attachmentUrl":  content.AttachmentURL,
		"author":         content.Author,
		"authorCode":     content.AuthorCode,
		"source":         content.Source,
		"publishTime":    formatLocalTime(content.PublishTime),
		"status":         content.Status,
		"isTop":          content.IsTop,
		"viewCount":      content.ViewCount,
		"createdAt":      content.CreatedAt.Format("2006-01-02 15:04:05"),
		"updatedAt":      content.UpdatedAt.Format("2006-01-02 15:04:05"),
	}))
}

func (c *MemberZoneController) CreateMemberContent(ctx *gin.Context) {
	var req models.MemberContent
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}
	// 作者信息一律以当前登录用户为准，忽略请求体中的 author/authorCode/viewCount
	userID := ctx.GetUint("userID")
	if user, err := c.userService.GetUserByID(userID); err == nil && user != nil {
		req.Author = user.Username
		req.AuthorCode = strconv.FormatUint(uint64(user.ID), 10)
	}
	req.ViewCount = 0

	if err := c.contentService.CreateMemberContent(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("发布会员内容失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("发布会员内容成功", nil))
}

func (c *MemberZoneController) UpdateMemberContent(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "内容ID无效"))
		return
	}
	content, err := c.contentService.GetMemberContentByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "内容不存在"))
		return
	}
	userID := ctx.GetUint("userID")
	if !c.canAccessMemberContent(content, userID) {
		ctx.JSON(http.StatusOK, utils.Error(1, "无权操作"))
		return
	}

	var req models.MemberContent
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}
	if err := c.contentService.UpdateMemberContent(uint(id), &req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("更新会员内容失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新会员内容成功", nil))
}

func (c *MemberZoneController) UpdateMemberContentStatus(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "内容ID无效"))
		return
	}
	content, err := c.contentService.GetMemberContentByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "内容不存在"))
		return
	}
	userID := ctx.GetUint("userID")
	if !c.canAccessMemberContent(content, userID) {
		ctx.JSON(http.StatusOK, utils.Error(1, "无权操作"))
		return
	}

	var req struct {
		Status int `json:"status"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误"))
		return
	}
	if err := c.contentService.UpdateMemberContentStatus(uint(id), req.Status); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("更新内容状态失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新内容状态成功", nil))
}

func (c *MemberZoneController) DeleteMemberContent(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "内容ID无效"))
		return
	}
	content, err := c.contentService.GetMemberContentByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "内容不存在"))
		return
	}
	userID := ctx.GetUint("userID")
	if !c.canAccessMemberContent(content, userID) {
		ctx.JSON(http.StatusOK, utils.Error(1, "无权操作"))
		return
	}
	if err := c.contentService.DeleteMemberContent(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("删除会员内容失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("删除会员内容成功", nil))
}
