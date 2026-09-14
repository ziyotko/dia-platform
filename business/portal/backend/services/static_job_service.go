package services

import (
	"context"
	"net/http"
	"strings"

	"server/utils"
)

// StaticJobService 静态化程序代理服务（业务侧复用）。
// 承载如「文章审核通过/发布后生成详情页静态文件」等由服务层触发的静态化调用，
// 全部为尽力而为（best-effort）：静态化参数未配置、程序不可达或调用失败时仅记录日志，不影响主流程。
type StaticJobService struct {
	settingsService  *SettingsService
	staticLogService *StaticLogService
}

func NewStaticJobService() *StaticJobService {
	return &StaticJobService{
		settingsService:  &SettingsService{},
		staticLogService: &StaticLogService{},
	}
}

// GenerateArticleStaticByID 生成指定文章ID的详情页静态文件（尽力而为）。
// 读取后端全局静态化参数发起 POST /api/static/article?id={文章ID}&path={静态化输出路径}。
// 参数未配置或调用失败时返回错误并仅记录日志，不影响主流程。
func (s *StaticJobService) GenerateArticleStaticByID(ctx context.Context, id string) (*StaticProgramResult, error) {
	if strings.TrimSpace(id) == "" {
		return nil, nil
	}
	params, err := s.settingsService.GetStaticParamsFromCache()
	if err != nil {
		utils.Logger.Warnf("生成文章[%s]静态页失败：获取静态化参数失败: %s", id, err)
		return nil, err
	}
	path := strings.TrimSpace(params.StaticPath)
	if path == "" {
		utils.Logger.Warnf("生成文章[%s]静态页失败：静态化输出路径未配置", id)
		return nil, &StaticProgramError{Kind: StaticErrConfigNotSet, Message: "静态化输出路径未配置"}
	}
	res, err := s.CallWithParams(ctx, params, http.MethodPost, "/api/static/article", map[string]string{
		"id":   id,
		"path": path,
	})
	if err != nil {
		return nil, err
	}
	// 与手动「重新生成」共用同一日志记录逻辑，保证静态化日志、最后静态化时间、今日文件数口径一致
	s.staticLogService.RecordPageStaticDone("系统", "生成详情页任务完成", id, res.StatusCode, res.Body)
	return res, nil
}
