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

// memberContentBriefDTO 列表精简字段（不含正文 longtext），供管理列表使用。
func memberContentBriefDTO(item models.MemberContent, columnName string) gin.H {
	return gin.H{
		"id":               item.ID,
		"memberColumnId":   item.MemberColumnID,
		"memberColumnName": columnName,
		"title":            item.Title,
		"type":             item.Type,
		"cover":            item.Cover,
		"dataYear":         item.DataYear,
		"unitName":         item.UnitName,
		"issueNo":          item.IssueNo,
		"publishYearMonth": item.PublishYearMonth,
		"author":           item.Author,
		"authorCode":       item.AuthorCode,
		"source":           item.Source,
		"publishTime":      formatLocalTime(item.PublishTime),
		"status":           item.Status,
		"isTop":            item.IsTop,
		"viewCount":        item.ViewCount,
		"createdAt":        item.CreatedAt.Format("2006-01-02 15:04:05"),
		"updatedAt":        item.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// memberContentFullDTO 内容完整字段（含各类型专属字段）。
// withContent=false 时不带正文（正文为 longtext，字段较多时体积大；对外栏目内容列表用）。
func memberContentFullDTO(item models.MemberContent, columnName string, withContent bool) gin.H {
	dto := gin.H{
		"id":                 item.ID,
		"memberColumnId":     item.MemberColumnID,
		"memberColumnName":   columnName,
		"title":              item.Title,
		"type":               item.Type,
		"source":             item.Source,
		"publishTime":        formatLocalTime(item.PublishTime),
		"status":             item.Status,
		"isTop":              item.IsTop,
		"cover":              item.Cover,
		"attachmentName":     item.AttachmentName,
		"attachmentUrl":      item.AttachmentURL,
		"dataYear":           item.DataYear,
		"unitName":           item.UnitName,
		"province":           item.Province,
		"region":             item.Region,
		"isBelt":             item.IsBelt,
		"isAxis":             item.IsAxis,
		"subField":           item.SubField,
		"mainBusinessIncome": item.MainBusinessIncome,
		"fullVideoUrl":       item.FullVideoURL,
		"previewVideoUrl":    item.PreviewVideoURL,
		"issueNo":            item.IssueNo,
		"publishYearMonth":   item.PublishYearMonth,
		"summary":            item.Summary,
		"paperFileName":      item.PaperFileName,
		"paperFileUrl":       item.PaperFileURL,
		"author":             item.Author,
		"authorCode":         item.AuthorCode,
		"viewCount":          item.ViewCount,
		"createdAt":          item.CreatedAt.Format("2006-01-02 15:04:05"),
		"updatedAt":          item.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	if withContent {
		dto["content"] = item.Content
	}
	return dto
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
		list = append(list, memberContentBriefDTO(item, columnNames[item.MemberColumnID]))
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
	columnNames, err := c.contentService.GetMemberColumnNames([]uint{content.MemberColumnID})
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取会员内容失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取会员内容成功",
		memberContentFullDTO(*content, columnNames[content.MemberColumnID], true)))
}

// ================= 对外只读接口（需登录，仅返回「已发布」内容） =================

// GetColumnMemberContents 对外接口：分页取指定会员栏目下「已发布」的内容（置顶优先）。
// GET /member-contents/column/:columnId?page=&pageSize= ；列表不含正文（正文请用下面的详情接口）。
func (c *MemberZoneController) GetColumnMemberContents(ctx *gin.Context) {
	columnID, err := strconv.ParseUint(ctx.Param("columnId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "会员栏目ID无效"))
		return
	}
	page := memberZoneQueryPage(ctx.DefaultQuery("page", "1"), 1)
	pageSize := memberZoneQueryPage(ctx.DefaultQuery("pageSize", "10"), 10)

	contents, total, err := c.contentService.GetPublishedMemberContentsByColumn(uint(columnID), page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("获取会员栏目内容失败", err)))
		return
	}
	columnNames, err := c.contentService.GetMemberColumnNames([]uint{uint(columnID)})
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取会员栏目内容失败"))
		return
	}
	columnName := columnNames[uint(columnID)]
	list := make([]gin.H, 0, len(contents))
	for _, item := range contents {
		list = append(list, memberContentFullDTO(item, columnName, false))
	}
	ctx.JSON(http.StatusOK, utils.Success("获取会员栏目内容成功", utils.PageData(list, total, page, pageSize)))
}

// GetMemberContentDetail 对外接口：按 ID 取「已发布」内容的完整信息（含正文）。
// GET /member-contents/detail/:id
func (c *MemberZoneController) GetMemberContentDetail(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "内容ID无效"))
		return
	}
	content, err := c.contentService.GetPublishedMemberContentByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("获取会员内容失败", err)))
		return
	}
	columnNames, err := c.contentService.GetMemberColumnNames([]uint{content.MemberColumnID})
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取会员内容失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取会员内容成功",
		memberContentFullDTO(*content, columnNames[content.MemberColumnID], true)))
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
