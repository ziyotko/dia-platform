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

// staticProgramHTTPTimeout 静态化程序 HTTP 调用超时（controller 与 service 共用，避免多处魔法数字）
const staticProgramHTTPTimeout = 120 * time.Second

// StaticProgramErrorKind 静态化调用失败类型
type StaticProgramErrorKind string

const (
	// StaticErrConfigNotSet 静态化参数未配置（输出路径/访问地址等）
	StaticErrConfigNotSet StaticProgramErrorKind = "config"
	// StaticErrBuildRequest 构造 HTTP 请求失败
	StaticErrBuildRequest StaticProgramErrorKind = "request"
	// StaticErrUnreachable 无法连接静态化程序
	StaticErrUnreachable StaticProgramErrorKind = "unreachable"
	// StaticErrReadBody 读取静态化程序响应体失败
	StaticErrReadBody StaticProgramErrorKind = "read"
)

// StaticProgramError 静态化程序调用错误，携带失败类型，便于调用方映射为不同响应。
type StaticProgramError struct {
	Kind    StaticProgramErrorKind
	Message string
	Err     error
}

func (e *StaticProgramError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *StaticProgramError) Unwrap() error { return e.Err }

// StaticProgramResult 静态化程序响应
type StaticProgramResult struct {
	StatusCode  int
	ContentType string
	Body        []byte
}

// staticProgramBaseURL 规范化静态化程序访问地址（缺失协议时默认 http，并去除末尾斜杠）
func staticProgramBaseURL(params *StaticParams) string {
	addr := strings.TrimSpace(params.StaticProgramAddr)
	if addr == "" {
		return ""
	}
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}
	return strings.TrimRight(addr, "/")
}

// staticProgramToken 获取静态化程序访问令牌。
// 配置项存放的是「环境变量名」（如 CAAM_STATIC_TOKEN），实际令牌值从该环境变量读取。
// 环境变量未配置时返回明确错误，而不是回退使用令牌名本身：
//   - 回退只会让上游返回 401，前端只能看到笼统的请求失败，排查时看不到真实原因；
//   - 若运维把真实令牌误填进「令牌名」字段，还会被当作令牌明文发往上游。
func staticProgramToken(params *StaticParams) (string, error) {
	name := strings.TrimSpace(params.StaticProgramTokenName)
	if name == "" {
		return "", nil
	}
	if v := os.Getenv(name); v != "" {
		return v, nil
	}
	return "", &StaticProgramError{
		Kind:    StaticErrConfigNotSet,
		Message: "静态化程序访问令牌未配置: 请先设置环境变量 " + name + "（该名称由「系统设置-静态化设置」的令牌名指定）",
	}
}

// Call 读取全局静态化参数并调用静态化程序。
func (s *StaticJobService) Call(ctx context.Context, method, targetPath string, query map[string]string) (*StaticProgramResult, error) {
	params, err := s.settingsService.GetStaticParamsFromCache()
	if err != nil {
		return nil, &StaticProgramError{Kind: StaticErrConfigNotSet, Message: "获取静态化参数失败", Err: err}
	}
	return s.CallWithParams(ctx, params, method, targetPath, query)
}

// CallWithParams 使用已解析的静态化参数调用静态化程序，避免在同一次请求内重复读取缓存。
func (s *StaticJobService) CallWithParams(ctx context.Context, params *StaticParams, method, targetPath string, query map[string]string) (*StaticProgramResult, error) {
	base := staticProgramBaseURL(params)
	if base == "" {
		return nil, &StaticProgramError{Kind: StaticErrConfigNotSet, Message: "静态化程序访问地址未配置，请先在「系统设置-静态化设置」中配置"}
	}

	target := base + targetPath
	req, err := http.NewRequestWithContext(ctx, method, target, nil)
	if err != nil {
		return nil, &StaticProgramError{Kind: StaticErrBuildRequest, Message: "构造静态化请求失败", Err: err}
	}
	token, err := staticProgramToken(params)
	if err != nil {
		return nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	q := req.URL.Query()
	for k, v := range query {
		q.Set(k, v)
	}
	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: staticProgramHTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		utils.Logger.Warnf("调用静态化程序失败: %s, url=%s", err, target)
		return nil, &StaticProgramError{Kind: StaticErrUnreachable, Message: "无法连接静态化程序，请检查「静态化程序访问地址」配置及服务运行状态", Err: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &StaticProgramError{Kind: StaticErrReadBody, Message: "读取静态化程序响应失败", Err: err}
	}

	utils.Logger.Infof("静态化任务请求: method=%s url=%s status=%d", method, target, resp.StatusCode)

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}
	return &StaticProgramResult{
		StatusCode:  resp.StatusCode,
		ContentType: contentType,
		Body:        body,
	}, nil
}
