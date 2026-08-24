package controllers

import (
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"server/services"
	"server/utils"
)

// StaticJobController 静态化批量操作任务控制器
// 四个批量操作（全站静态化/生成首页/生成栏目页/生成详情页）均由本控制器接收请求，
// 读取后端全局静态化参数（静态化输出路径/静态化程序访问地址/静态化程序访问令牌名/首页整体变灰），
// 然后代理转发给静态化程序，并将静态化程序返回的 202 状态码与任务信息原样透传给前端。
type StaticJobController struct {
	settingsService *services.SettingsService
}

func NewStaticJobController() *StaticJobController {
	return &StaticJobController{
		settingsService: &services.SettingsService{},
	}
}

// staticProgramBaseURL 规范化静态化程序访问地址（缺失协议时默认 http，并去除末尾斜杠）
func (c *StaticJobController) staticProgramBaseURL(params *services.StaticParams) string {
	addr := strings.TrimSpace(params.StaticProgramAddr)
	if addr == "" {
		return ""
	}
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}
	return strings.TrimRight(addr, "/")
}

// staticProgramToken 获取静态化程序访问令牌（全局变量）
// 静态化程序访问令牌名配置项存放的是环境变量（全局变量）的名称，例如 CAAM_STATIC_TOKEN；
// 实际令牌值从该环境变量中读取。若环境变量未配置，则回退直接使用令牌名本身作为令牌。
func (c *StaticJobController) staticProgramToken(params *services.StaticParams) string {
	name := strings.TrimSpace(params.StaticProgramTokenName)
	if name == "" {
		return ""
	}
	if v := os.Getenv(name); v != "" {
		return v
	}
	return name
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

// proxyToStaticProgram 代理转发请求到静态化程序，并原样透传状态码与响应体
func (c *StaticJobController) proxyToStaticProgram(ctx *gin.Context, params *services.StaticParams, method, targetPath string, query map[string]string) {
	base := c.staticProgramBaseURL(params)
	if base == "" {
		ctx.JSON(http.StatusOK, utils.Error(1, "静态化程序访问地址未配置，请先在「基础配置-静态化设置」中配置"))
		return
	}

	target := base + targetPath
	req, err := http.NewRequestWithContext(ctx.Request.Context(), method, target, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.Error(1, "构造静态化请求失败: "+err.Error()))
		return
	}

	// 请求头增加 Authorization: Bearer 静态化程序访问令牌
	if token := c.staticProgramToken(params); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// 组装查询参数
	q := req.URL.Query()
	for k, v := range query {
		q.Set(k, v)
	}
	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		utils.Logger.Warnf("调用静态化程序失败: %s, url=%s", err, target)
		ctx.JSON(http.StatusBadGateway, utils.Error(1, "无法连接静态化程序，请检查「静态化程序访问地址」配置及服务运行状态"))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, utils.Error(1, "读取静态化程序响应失败: "+err.Error()))
		return
	}

	utils.Logger.Infof("静态化任务请求: method=%s url=%s status=%d", method, target, resp.StatusCode)

	// 原样透传状态码与响应体：
	// - 发起任务（site/pages/lists/articles）成功返回 202 及任务信息，代表请求已发送，请等待处理结果
	// - 查询任务状态（jobs/:id）返回 200 及任务状态结构
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}
	ctx.Data(resp.StatusCode, contentType, body)
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
	c.proxyToStaticProgram(ctx, params, http.MethodPost, "/api/static/site", map[string]string{
		"path": path,
		"gray": c.resolveGray(ctx, params),
	})
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
	c.proxyToStaticProgram(ctx, params, http.MethodPost, "/api/static/pages", map[string]string{
		"path": path,
		"gray": c.resolveGray(ctx, params),
	})
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
	c.proxyToStaticProgram(ctx, params, http.MethodPost, "/api/static/lists", map[string]string{
		"path": path,
	})
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
	c.proxyToStaticProgram(ctx, params, http.MethodPost, "/api/static/articles", map[string]string{
		"path": path,
	})
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
	c.proxyToStaticProgram(ctx, params, http.MethodGet, "/api/static/jobs/"+id, nil)
}
