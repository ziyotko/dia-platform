package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"portal/config"
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

// memberZoneMaxPageSize 单页条数上限：
// 防止外部系统/前端传超大 pageSize 一次拉全表（拖慢数据库、打满内存）。
// 与前端分页组件的可选项 [10, 20, 50, 100] 对齐，故管理页面不受影响。
const memberZoneMaxPageSize = 100

// memberZoneQueryPageSize 解析 pageSize：< 1 回退缺省值（10），超过上限按上限截断，
// 并保证与 utils.PageData 回显的 pageSize 一致。
func memberZoneQueryPageSize(raw string, fallback int) int {
	v := memberZoneQueryInt(raw, fallback)
	if v < 1 {
		return fallback
	}
	if v > memberZoneMaxPageSize {
		return memberZoneMaxPageSize
	}
	return v
}

// ============================ 会员栏目（会员栏目分类） ============================

func (c *MemberZoneController) GetMemberColumns(ctx *gin.Context) {
	name := ctx.Query("name")
	status := memberZoneQueryInt(ctx.Query("status"), -1)
	page := memberZoneQueryPage(ctx.DefaultQuery("page", "1"), 1)
	pageSize := memberZoneQueryPageSize(ctx.DefaultQuery("pageSize", "10"), 10)

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
	// 所属会员栏目的启用状态：-1 全部 / 0 禁用 / 1 启用
	columnStatus := memberZoneQueryInt(ctx.Query("columnStatus"), -1)
	page := memberZoneQueryPage(ctx.DefaultQuery("page", "1"), 1)
	pageSize := memberZoneQueryPageSize(ctx.DefaultQuery("pageSize", "10"), 10)

	userID := ctx.GetUint("userID")
	authorCodeScope := ""
	if !models.HasAdminRoleIDs(c.userService.MustGetUserRoleIds(userID)) {
		authorCodeScope = strconv.FormatUint(uint64(userID), 10)
	}

	contents, total, err := c.contentService.GetMemberContents(services.MemberContentQuery{
		Title:        title,
		ColumnID:     columnID,
		Type:         contentType,
		Status:       status,
		ColumnStatus: columnStatus,
		// 发布时间区间（YYYY-MM-DD，含当天，两端可只传一端）
		PublishStart:    ctx.Query("publishStart"),
		PublishEnd:      ctx.Query("publishEnd"),
		AuthorCodeScope: authorCodeScope,
		Page:            page,
		PageSize:        pageSize,
	})
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
// GET /member-contents/column/:key?page=&pageSize=
// `:key` 为「会员栏目标识」：纯数字按 ID 匹配，其它按名称精确匹配（名称唯一）。
// 这样外部系统（如 CAMIE）按名称调用即可跨环境使用，无需维护「ID 映射配置」。
// 列表不含正文（正文请用下面的详情接口）。
func (c *MemberZoneController) GetColumnMemberContents(ctx *gin.Context) {
	key := strings.TrimSpace(ctx.Param("key"))
	column, err := c.columnService.GetMemberColumnByKey(key)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("获取会员栏目内容失败", err)))
		return
	}
	page := memberZoneQueryPage(ctx.DefaultQuery("page", "1"), 1)
	pageSize := memberZoneQueryPageSize(ctx.DefaultQuery("pageSize", "10"), 10)

	contents, total, err := c.contentService.GetPublishedMemberContentsByColumn(column.ID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("获取会员栏目内容失败", err)))
		return
	}
	list := make([]gin.H, 0, len(contents))
	for _, item := range contents {
		list = append(list, memberContentFullDTO(item, column.Name, false))
	}
	ctx.JSON(http.StatusOK, utils.Success("获取会员栏目内容成功", utils.PageData(list, total, page, pageSize)))
}

// GetMemberColumnOptions 对外接口：返回全部会员栏目（含禁用）的 id/名称/状态，
// 供外部系统确认栏目写法，或自行按名称→ID 映射调用。
// GET /member-columns/options
func (c *MemberZoneController) GetMemberColumnOptions(ctx *gin.Context) {
	columns, err := c.columnService.GetAllMemberColumns()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取会员栏目失败"))
		return
	}
	list := make([]gin.H, 0, len(columns))
	for _, item := range columns {
		list = append(list, gin.H{
			"id":     item.ID,
			"name":   item.Name,
			"status": item.Status,
		})
	}
	ctx.JSON(http.StatusOK, utils.Success("获取会员栏目成功", utils.AllData(list, int64(len(list)))))
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

// ================= 会员专区文件（私有目录 + 短时效签名 URL） =================

// memberFileNamePattern 会员专区私有文件的合法文件名：
// `<纳秒时间戳>_<16位随机hex>.<扩展名>`（由 upload_controller 生成）。
// 严格限定字符集与长度，杜绝 `../`、绝对路径等路径穿越写法。
var memberFileNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,80}\.[A-Za-z0-9]{1,10}$`)

// memberFileSignedURL 拼出带签名的文件访问地址（相对地址，与 VITE_API_BASE_URL 同源前缀）。
func memberFileSignedURL(name string, exp int64, nonce string, sign string) string {
	return fmt.Sprintf("%s/member-files/%s?exp=%d&nonce=%s&sign=%s",
		config.AppConfig.Server.ApiPrefix, name, exp, nonce, sign)
}

// SignMemberFile 下发会员专区文件的「短时效签名 URL」（需登录）。
// GET /member-files/sign?name=<文件名>
//
// 访问判定（按文件名反查 member_content，见 FindMemberFileReferences）：
//   - 被【已发布】内容引用 → 任何登录用户都可取签名地址（会员内容面向已登录会员）；
//   - 仅被【本人】的内容引用（草稿/已下线） → 作者本人可取（编辑器需要预览草稿的封面/附件）；
//   - 仅被【他人未发布】内容引用 → 仅管理员可取；
//   - 完全没有被任何内容引用（如上传后尚未保存） → 拒绝；这种情况请直接用上传接口当场返回的 signedUrl。
func (c *MemberZoneController) SignMemberFile(ctx *gin.Context) {
	name := strings.TrimSpace(ctx.Query("name"))
	if !memberFileNamePattern.MatchString(name) {
		ctx.JSON(http.StatusOK, utils.Error(1, "文件不存在"))
		return
	}
	refs, err := c.contentService.FindMemberFileReferences(name)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取文件访问地址失败"))
		return
	}
	userID := ctx.GetUint("userID")
	isAdmin := models.HasAdminRoleIDs(c.userService.MustGetUserRoleIds(userID))
	authorCode := strconv.FormatUint(uint64(userID), 10)

	allowed := false
	for _, ref := range refs {
		if ref.Status == models.MemberContentStatusPublished || isAdmin || ref.AuthorCode == authorCode {
			allowed = true
			break
		}
	}
	if !allowed {
		ctx.JSON(http.StatusOK, utils.Error(1, "无权访问该文件"))
		return
	}

	exp, nonce, sign := utils.SignFileAccess(name, utils.FileAccessSignTTL)
	ctx.JSON(http.StatusOK, utils.Success("获取文件访问地址成功", gin.H{
		"url":       memberFileSignedURL(name, exp, nonce, sign),
		"expiresIn": int(utils.FileAccessSignTTL.Seconds()),
	}))
}

// GetMemberFile 会员专区文件读取接口：**凭短时效签名访问**（签名即凭证，不再要求登录态）。
// GET /member-files/:name?exp=&nonce=&sign=
//
// 为什么是签名而不是 Token：
//   - `<img src>` / `<video src>` / `<a href>` 无法设置 `Authorization` 头；
//   - 把 Token 放在 URL 里会进入访问日志/Referer（且是整个会话凭证，泄露代价大）；
//   - 签名地址由 GET /member-files/sign 下发（那里做了「引用 + 身份」判定）、**5 分钟过期**，
//     即使 URL 外泄也只能短暂读取单个文件。
func (c *MemberZoneController) GetMemberFile(ctx *gin.Context) {
	name := ctx.Param("name")
	if !memberFileNamePattern.MatchString(name) {
		ctx.JSON(http.StatusOK, utils.Error(1, "文件不存在"))
		return
	}
	exp, _ := strconv.ParseInt(ctx.Query("exp"), 10, 64)
	if err := utils.VerifyFileAccessSign(name, exp, ctx.Query("nonce"), ctx.Query("sign")); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("读取文件失败", err)))
		return
	}

	f, err := os.Open(filepath.Join(memberPrivateUploadDir, name))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "文件不存在"))
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || info.IsDir() {
		ctx.JSON(http.StatusOK, utils.Error(1, "文件不存在"))
		return
	}
	// ServeContent 会根据扩展名/MIME 嗅探写入 Content-Type，并原生支持 Range
	// （大视频拖动进度条、断点续传都依赖它）与 If-Modified-Since。
	http.ServeContent(ctx.Writer, ctx.Request, name, info.ModTime(), f)
}
