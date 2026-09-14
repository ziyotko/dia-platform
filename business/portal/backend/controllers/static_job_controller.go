package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/services"
	"server/utils"
)

// StaticJobController 静态化批量操作任务控制器
// 四个批量操作（生成全站/生成首页/生成栏目页/生成详情页）均由本控制器接收请求，
// 读取后端全局静态化参数（静态化输出路径/静态化程序访问地址/静态化程序访问令牌名/首页整体变灰），
// 然后代理转发给静态化程序，并将静态化程序返回的 202 状态码与任务信息原样透传给前端。
type StaticJobController struct {
	settingsService  *services.SettingsService
	staticLogService *services.StaticLogService
	staticJobService *services.StaticJobService
}

func NewStaticJobController() *StaticJobController {
	return &StaticJobController{
		settingsService:  &services.SettingsService{},
		staticLogService: &services.StaticLogService{},
		staticJobService: services.NewStaticJobService(),
	}
}

// 静态化程序返回的任务结构（用于透传响应解析与日志记录）
type staticProgramJob struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	StartedAt string `json:"started_at"`
	Progress  struct {
		Stage          string `json:"stage"`
		Processed      int    `json:"processed"`
		Total          int    `json:"total"`
		GeneratedFiles int    `json:"generated_files"`
	} `json:"progress"`
}

// 静态化程序返回的统一响应结构
type staticProgramResponse struct {
	OK  bool              `json:"ok"`
	Job *staticProgramJob `json:"job"`
}

// 任务类型 → 中文名
var staticKindName = map[string]string{
	"site":     "生成全站",
	"pages":    "生成首页",
	"lists":    "生成栏目页",
	"articles": "生成详情页",
}

func staticKindText(kind string) string {
	if name, ok := staticKindName[kind]; ok {
		return name
	}
	return kind
}

// 组装任务日志完整内容（任务类型/任务ID/状态/阶段/进度/生成文件数）
func staticJobLogMessage(job *staticProgramJob, statusText string) string {
	parts := []string{
		"任务类型：" + staticKindText(job.Kind),
		"任务ID：" + job.ID,
		"状态：" + statusText,
	}
	if job.Progress.Stage != "" {
		parts = append(parts, "阶段："+job.Progress.Stage)
	}
	if job.Progress.Total > 0 {
		parts = append(parts, fmt.Sprintf("已处理 %d / %d", job.Progress.Processed, job.Progress.Total))
	}
	if job.Progress.GeneratedFiles > 0 {
		parts = append(parts, fmt.Sprintf("已生成文件 %d", job.Progress.GeneratedFiles))
	}
	return strings.Join(parts, "｜")
}

// 计算任务耗时：由 started_at / updated_at 计算，格式化为 时/分/秒
func staticJobDuration(job *staticProgramJob) string {
	if job.StartedAt == "" {
		return "-"
	}
	start, err := time.Parse(time.RFC3339Nano, job.StartedAt)
	if err != nil {
		return "-"
	}
	end := start
	if job.UpdatedAt != "" {
		if t, err := time.Parse(time.RFC3339Nano, job.UpdatedAt); err == nil {
			end = t
		}
	}
	total := max(int64(end.Sub(start).Seconds()), 0)
	h, m, s := total/3600, (total%3600)/60, total%60
	switch {
	case h > 0:
		return fmt.Sprintf("%d时%d分%d秒", h, m, s)
	case m > 0:
		return fmt.Sprintf("%d分%d秒", m, s)
	default:
		return fmt.Sprintf("%d秒", s)
	}
}

// 当前操作人：优先用户名，其次账号/邮箱，默认 admin
func (c *StaticJobController) currentOperator(ctx *gin.Context) string {
	uid := ctx.GetUint("userID")
	if uid > 0 {
		var user models.User
		if err := utils.DB.First(&user, uid).Error; err == nil {
			if user.Username != "" {
				return user.Username
			}
			if user.Account != "" {
				return user.Account
			}
			if user.Email != "" {
				return user.Email
			}
		}
	}
	return "admin"
}

// writeStaticLog 记录静态化日志（后端业务统一记录，按 任务ID+状态 自动去重）
func (c *StaticJobController) writeStaticLog(ctx *gin.Context, status, operation, message string, job *staticProgramJob) {
	log := &models.StaticLog{
		Operation: operation,
		PageName:  "-",
		Path:      "-",
		Duration:  staticJobDuration(job),
		FileSize:  "-",
		Operator:  c.currentOperator(ctx),
		Status:    status,
		Message:   message,
		JobID:     job.ID,
	}
	if job.Progress.GeneratedFiles > 0 {
		log.FileSize = fmt.Sprintf("%d 个文件", job.Progress.GeneratedFiles)
	}
	if err := c.staticLogService.CreateIfNotExists(log); err != nil {
		utils.Logger.Warnf("记录静态化日志失败: %s", err)
	}
}

// writeTaskSubmitLog 任务提交成功后记录静态化日志（仅 202 且响应含任务信息时记录）
func (c *StaticJobController) writeTaskSubmitLog(ctx *gin.Context, statusCode int, body []byte, kind string) {
	if statusCode != http.StatusAccepted {
		return
	}
	var resp staticProgramResponse
	if err := json.Unmarshal(body, &resp); err != nil || !resp.OK || resp.Job == nil {
		return
	}
	c.writeStaticLog(ctx, "primary", staticKindText(kind)+"任务提交", staticJobLogMessage(resp.Job, "等待执行"), resp.Job)
}

// writeTaskDoneLog 任务进入终态后记录完成/失败/中断日志（由查询任务状态接口检测）
func (c *StaticJobController) writeTaskDoneLog(ctx *gin.Context, statusCode int, body []byte) {
	if statusCode != http.StatusOK {
		return
	}
	var resp staticProgramResponse
	if err := json.Unmarshal(body, &resp); err != nil || !resp.OK || resp.Job == nil {
		return
	}
	job := resp.Job
	switch job.Status {
	case "succeeded":
		c.writeStaticLog(ctx, "success", staticKindText(job.Kind)+"任务完成", staticJobLogMessage(job, "执行成功"), job)
	case "failed":
		c.writeStaticLog(ctx, "danger", staticKindText(job.Kind)+"任务失败", staticJobLogMessage(job, "执行失败"), job)
	case "interrupted":
		c.writeStaticLog(ctx, "warning", staticKindText(job.Kind)+"任务中断", staticJobLogMessage(job, "已中断"), job)
	}
}

// resolveStaticParams 读取静态化参数（优先缓存，缓存未命中回源数据库）
func (c *StaticJobController) resolveStaticParams(ctx *gin.Context) (*services.StaticParams, bool) {
	params, err := c.settingsService.GetStaticParamsFromCache()
	if err != nil {
		utils.Logger.Warnf("获取静态化参数失败: %s", err)
		ctx.JSON(http.StatusOK, utils.Error(1, "获取静态化参数失败"))
		return nil, false
	}
	return params, true
}

// resolveOutputPath 输出目录：优先使用请求显式传入的 path，否则读取后端全局变量「静态化输出路径」
func (c *StaticJobController) resolveOutputPath(ctx *gin.Context, params *services.StaticParams) (string, bool) {
	if p := strings.TrimSpace(ctx.Query("path")); p != "" {
		return p, true
	}
	if p := strings.TrimSpace(params.StaticPath); p != "" {
		return p, true
	}
	ctx.JSON(http.StatusOK, utils.Error(1, "静态化输出路径未配置，请先在「基础配置-静态化设置」中配置"))
	return "", false
}

// resolveGray 首页整体变灰：请求显式传入 gray（1 开启 / 2 关闭）优先，否则读取后端全局变量「首页整体变灰」
func (c *StaticJobController) resolveGray(ctx *gin.Context, params *services.StaticParams) string {
	if g := strings.TrimSpace(ctx.Query("gray")); g == "1" || g == "2" {
		return g
	}
	if params.HomeGray {
		return "1"
	}
	return "2"
}

// proxyToStaticProgram 代理转发请求到静态化程序，返回状态码与响应体（调用方负责写回客户端并记录静态化日志）
func (c *StaticJobController) proxyToStaticProgram(ctx *gin.Context, params *services.StaticParams, method, targetPath string, query map[string]string) (int, []byte) {
	res, err := c.staticJobService.CallWithParams(ctx.Request.Context(), params, method, targetPath, query)
	if err != nil {
		var progErr *services.StaticProgramError
		if errors.As(err, &progErr) {
			switch progErr.Kind {
			case services.StaticErrConfigNotSet:
				// 参数未配置属业务提示，统一 200 + code=1
				ctx.JSON(http.StatusOK, utils.Error(1, progErr.Message))
			case services.StaticErrBuildRequest:
				ctx.JSON(http.StatusInternalServerError, utils.Error(1, utils.SanitizeError(progErr.Message, progErr.Err)))
			case services.StaticErrReadBody:
				ctx.JSON(http.StatusBadGateway, utils.Error(1, utils.SanitizeError(progErr.Message, progErr.Err)))
			default:
				// 无法连接静态化程序：保持既有 502 + code=1 透传
				ctx.JSON(http.StatusBadGateway, utils.Error(1, progErr.Message))
			}
		} else {
			ctx.JSON(http.StatusBadGateway, utils.Error(1, "静态化服务调用失败"))
		}
		return 0, nil
	}

	// 透传状态码与响应体：
	// - 发起任务（site/pages/lists/articles）成功返回 202 及任务信息，代表请求已发送，请等待处理结果
	// - 查询任务状态（jobs/:id）返回 200 及任务状态结构
	ctx.Data(res.StatusCode, res.ContentType, res.Body)
	return res.StatusCode, res.Body
}

// SiteStatic 全站静态化：POST /static/site?path={输出目录}&gray={1或2}
func (c *StaticJobController) SiteStatic(ctx *gin.Context) {
	params, ok := c.resolveStaticParams(ctx)
	if !ok {
		return
	}
	path, ok := c.resolveOutputPath(ctx, params)
	if !ok {
		return
	}
	statusCode, body := c.proxyToStaticProgram(ctx, params, http.MethodPost, "/api/static/site", map[string]string{
		"path": path,
		"gray": c.resolveGray(ctx, params),
	})
	c.writeTaskSubmitLog(ctx, statusCode, body, "site")
}

// PagesStatic 生成首页：POST /static/pages?path={输出目录}&gray={1或2}
func (c *StaticJobController) PagesStatic(ctx *gin.Context) {
	params, ok := c.resolveStaticParams(ctx)
	if !ok {
		return
	}
	path, ok := c.resolveOutputPath(ctx, params)
	if !ok {
		return
	}
	statusCode, body := c.proxyToStaticProgram(ctx, params, http.MethodPost, "/api/static/pages", map[string]string{
		"path": path,
		"gray": c.resolveGray(ctx, params),
	})
	c.writeTaskSubmitLog(ctx, statusCode, body, "pages")
}

// ListsStatic 生成栏目页：POST /static/lists?path={输出目录}
func (c *StaticJobController) ListsStatic(ctx *gin.Context) {
	params, ok := c.resolveStaticParams(ctx)
	if !ok {
		return
	}
	path, ok := c.resolveOutputPath(ctx, params)
	if !ok {
		return
	}
	statusCode, body := c.proxyToStaticProgram(ctx, params, http.MethodPost, "/api/static/lists", map[string]string{
		"path": path,
	})
	c.writeTaskSubmitLog(ctx, statusCode, body, "lists")
}

// ArticlesStatic 生成详情页：POST /static/articles?path={输出目录}
func (c *StaticJobController) ArticlesStatic(ctx *gin.Context) {
	params, ok := c.resolveStaticParams(ctx)
	if !ok {
		return
	}
	path, ok := c.resolveOutputPath(ctx, params)
	if !ok {
		return
	}
	statusCode, body := c.proxyToStaticProgram(ctx, params, http.MethodPost, "/api/static/articles", map[string]string{
		"path": path,
	})
	c.writeTaskSubmitLog(ctx, statusCode, body, "articles")
}

// PageStatic 首页重新生成：POST /static/page?name={页面名}
// 输出目录与首页整体变灰均读取后端全局变量（静态化输出路径 / 首页整体变灰），
// 代理转发至静态化程序（自动附带 Authorization 验证令牌头），同步返回 HTTP 200 及生成结果。
func (c *StaticJobController) PageStatic(ctx *gin.Context) {
	params, ok := c.resolveStaticParams(ctx)
	if !ok {
		return
	}
	name := strings.TrimSpace(ctx.Query("name"))
	if name == "" {
		ctx.JSON(http.StatusOK, utils.Error(1, "页面名不能为空"))
		return
	}
	path, ok := c.resolveOutputPath(ctx, params)
	if !ok {
		return
	}
	statusCode, body := c.proxyToStaticProgram(ctx, params, http.MethodPost, "/api/static/page", map[string]string{
		"name": name,
		"path": path,
		"gray": c.resolveGray(ctx, params),
	})
	c.staticLogService.RecordPageStaticDone(c.currentOperator(ctx), "生成首页任务完成", name, statusCode, body)
}

// ListStatic 栏目页重新生成：POST /static/list?column_name={栏目名称}
// 输出目录读取后端全局变量（静态化输出路径），
// 代理转发至静态化程序（自动附带 Authorization 验证令牌头），同步返回 HTTP 200 及生成结果。
func (c *StaticJobController) ListStatic(ctx *gin.Context) {
	params, ok := c.resolveStaticParams(ctx)
	if !ok {
		return
	}
	columnName := strings.TrimSpace(ctx.Query("column_name"))
	if columnName == "" {
		ctx.JSON(http.StatusOK, utils.Error(1, "栏目名称不能为空"))
		return
	}
	path, ok := c.resolveOutputPath(ctx, params)
	if !ok {
		return
	}
	statusCode, body := c.proxyToStaticProgram(ctx, params, http.MethodPost, "/api/static/list", map[string]string{
		"column_name": columnName,
		"path":        path,
	})
	c.staticLogService.RecordPageStaticDone(c.currentOperator(ctx), "生成栏目页任务完成", columnName, statusCode, body)
}

// ArticleStatic 详情页重新生成：POST /static/article?id={文章ID}
// 输出目录读取后端全局变量（静态化输出路径），
// 代理转发至静态化程序（自动附带 Authorization 验证令牌头），同步返回 HTTP 200 及生成结果。
func (c *StaticJobController) ArticleStatic(ctx *gin.Context) {
	params, ok := c.resolveStaticParams(ctx)
	if !ok {
		return
	}
	id := strings.TrimSpace(ctx.Query("id"))
	if id == "" {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章ID不能为空"))
		return
	}
	path, ok := c.resolveOutputPath(ctx, params)
	if !ok {
		return
	}
	statusCode, body := c.proxyToStaticProgram(ctx, params, http.MethodPost, "/api/static/article", map[string]string{
		"id":   id,
		"path": path,
	})
	c.staticLogService.RecordPageStaticDone(c.currentOperator(ctx), "生成详情页任务完成", id, statusCode, body)
}

// DeleteArticleStatic 删除详情页静态文件：DELETE /static/article?id={文章ID}&path={输出目录}
// 输出目录优先取请求显式传入的 path，否则读取后端全局变量（静态化输出路径），
// 代理转发至静态化程序（自动附带 Authorization 验证令牌头），同步返回 HTTP 200 及删除结果。
func (c *StaticJobController) DeleteArticleStatic(ctx *gin.Context) {
	params, ok := c.resolveStaticParams(ctx)
	if !ok {
		return
	}
	id := strings.TrimSpace(ctx.Query("id"))
	if id == "" {
		ctx.JSON(http.StatusOK, utils.Error(1, "文章ID不能为空"))
		return
	}
	path, ok := c.resolveOutputPath(ctx, params)
	if !ok {
		return
	}
	statusCode, body := c.proxyToStaticProgram(ctx, params, http.MethodDelete, "/api/static/article", map[string]string{
		"id":   id,
		"path": path,
	})
	c.staticLogService.RecordArticleDeleteLog(c.currentOperator(ctx), id, statusCode, body)
}

// DeleteArticleStaticByID 删除指定文章ID的详情页静态文件（供文章下线/删除文章等业务复用）。
// 该调用为尽力而为（best-effort）：读取后端全局静态化参数发起删除，静态化参数未配置、
// 程序不可达或调用失败时仅记录日志，不写客户端响应，不影响主流程。
func (c *StaticJobController) DeleteArticleStaticByID(ctx *gin.Context, id string) {
	if strings.TrimSpace(id) == "" {
		return
	}
	params, err := c.settingsService.GetStaticParamsFromCache()
	if err != nil {
		utils.Logger.Warnf("删除文章[%s]静态文件失败：获取静态化参数失败: %s", id, err)
		return
	}
	path := strings.TrimSpace(params.StaticPath)
	if path == "" {
		utils.Logger.Warnf("删除文章[%s]静态文件失败：静态化输出路径未配置", id)
		return
	}

	res, err := c.staticJobService.CallWithParams(ctx.Request.Context(), params, http.MethodDelete, "/api/static/article", map[string]string{
		"id":   id,
		"path": path,
	})
	if err != nil {
		utils.Logger.Warnf("删除文章[%s]静态文件失败: %s", id, err)
		return
	}

	c.staticLogService.RecordArticleDeleteLog(c.currentOperator(ctx), id, res.StatusCode, res.Body)
}

// GetJob 查询任务状态：GET /static/jobs/{任务ID}
// 返回任务状态结构：queued 等待执行 / running 正在执行 / succeeded 执行成功 / failed 执行失败 / interrupted 被取消或超时中断
func (c *StaticJobController) GetJob(ctx *gin.Context) {
	params, ok := c.resolveStaticParams(ctx)
	if !ok {
		return
	}
	id := strings.TrimSpace(ctx.Param("id"))
	if id == "" {
		ctx.JSON(http.StatusOK, utils.Error(1, "任务ID不能为空"))
		return
	}
	statusCode, body := c.proxyToStaticProgram(ctx, params, http.MethodGet, "/api/static/jobs/"+id, nil)
	c.writeTaskDoneLog(ctx, statusCode, body)
}
