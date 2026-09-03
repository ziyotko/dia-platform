package services

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"server/utils"
)

// StaticJobService 静态化程序代理服务（业务侧复用）。
// 承载如「文章审核通过/发布后生成详情页静态文件」等由服务层触发的静态化调用，
// 全部为尽力而为（best-effort）：静态化参数未配置、程序不可达或调用失败时仅记录日志，不影响主流程。
type StaticJobService struct {
	settingsService *SettingsService
}

func NewStaticJobService() *StaticJobService {
	return &StaticJobService{
		settingsService: &SettingsService{},
	}
}

// staticProgramBaseURL 规范化静态化程序访问地址（缺失协议时默认 http，并去除末尾斜杠）
func (s *StaticJobService) staticProgramBaseURL(params *StaticParams) string {
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
func (s *StaticJobService) staticProgramToken(params *StaticParams) string {
	name := strings.TrimSpace(params.StaticProgramTokenName)
	if name == "" {
		return ""
	}
	if v := os.Getenv(name); v != "" {
		return v
	}
	return name
}

// GenerateArticleStaticByID 生成指定文章ID的详情页静态文件（尽力而为）。
// 读取后端全局静态化参数发起 POST /api/static/article?id={文章ID}&path={静态化输出路径}。
// 返回 HTTP 状态码与响应体；参数未配置或调用失败时返回 (0, nil) 并仅记录日志。
func (s *StaticJobService) GenerateArticleStaticByID(ctx context.Context, id string) (int, []byte) {
	if strings.TrimSpace(id) == "" {
		return 0, nil
	}
	params, err := s.settingsService.GetStaticParamsFromCache()
	if err != nil {
		utils.Logger.Warnf("生成文章[%s]静态页失败：获取静态化参数失败: %s", id, err)
		return 0, nil
	}
	base := s.staticProgramBaseURL(params)
	if base == "" {
		utils.Logger.Warnf("生成文章[%s]静态页失败：静态化程序访问地址未配置", id)
		return 0, nil
	}
	path := strings.TrimSpace(params.StaticPath)
	if path == "" {
		utils.Logger.Warnf("生成文章[%s]静态页失败：静态化输出路径未配置", id)
		return 0, nil
	}

	target := base + "/api/static/article"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, nil)
	if err != nil {
		utils.Logger.Warnf("构造静态化生成请求失败: %s", err)
		return 0, nil
	}
	if token := s.staticProgramToken(params); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	q := req.URL.Query()
	q.Set("id", id)
	q.Set("path", path)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		utils.Logger.Warnf("调用静态化程序生成文章[%s]静态页失败: %s, url=%s", id, err, target)
		return 0, nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	utils.Logger.Infof("生成文章静态页: id=%s url=%s status=%d", id, target, resp.StatusCode)
	return resp.StatusCode, body
}
