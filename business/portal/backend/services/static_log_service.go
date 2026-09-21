package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"server/models"
	"server/utils"
)

type StaticLogService struct{}

type StaticLogListResult struct {
	Total int64              `json:"total"`
	List  []models.StaticLog `json:"list"`
}

// StaticLatestTimes 各类型页面最后一次静态化成功时间（读取静态化日志最新成功记录）。
// 专题页既有批量任务（生成专题页）也有单页操作（重新生成/删除静态文件），
// 批量任务完成与单页同步生成共用「生成专题页任务完成」日志，故与首页/栏目页/详情页口径一致。
type StaticLatestTimes struct {
	Site       string `json:"site"`       // 全站最后静态化时间
	Home       string `json:"home"`       // 首页最后静态化时间
	Column     string `json:"column"`     // 栏目页最后静态化时间
	Topic      string `json:"topic"`      // 专题页最后静态化时间
	Detail     string `json:"detail"`     // 详情页最后静态化时间
	TodayFiles int    `json:"todayFiles"` // 今日静态化成功生成文件数
}

func (s *StaticLogService) GetList(page, pageSize int) (*StaticLogListResult, error) {
	var logs []models.StaticLog
	var total int64

	query := utils.DB.Model(&models.StaticLog{})

	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	err = query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	if err != nil {
		return nil, err
	}

	return &StaticLogListResult{
		Total: total,
		List:  logs,
	}, nil
}

// Clear 与操作日志/登录日志保持同一语义：仅清空半年前的静态化日志，保留最近半年。返回删除条数。
func (s *StaticLogService) Clear() (int64, error) {
	cutoff := time.Now().AddDate(0, -6, 0)
	result := utils.DB.Where("created_at < ?", cutoff).Delete(&models.StaticLog{})
	return result.RowsAffected, result.Error
}

// CreateIfNotExists 按 任务ID + 状态 去重创建静态化日志（同一任务同一状态只记录一次）
func (s *StaticLogService) CreateIfNotExists(log *models.StaticLog) error {
	if log.JobID != "" {
		var count int64
		if err := utils.DB.Model(&models.StaticLog{}).
			Where("job_id = ? AND status = ?", log.JobID, log.Status).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
	}
	return utils.DB.Create(log).Error
}

// StaticPageResult 静态化程序「单页/详情页同步生成」结果结构
type StaticPageResult struct {
	GeneratedAt      string  `json:"generated_at"`
	DurationSeconds  float64 `json:"duration_seconds"`
	Page             string  `json:"page"`
	TotalItems       int     `json:"total_items"`
	GeneratedFiles   int     `json:"generated_files"`
	GeneratedDetails int     `json:"generated_details"`
	GeneratedLists   int     `json:"generated_lists"`
	Output           string  `json:"output"`
	Gray             string  `json:"gray"`
}

// StaticPageResponse 单页同步生成统一响应结构
type StaticPageResponse struct {
	OK     bool              `json:"ok"`
	Result *StaticPageResult `json:"result"`
}

// RecordPageStaticDone 记录单页/详情页同步静态化成功日志。
// 手动「重新生成」与文章发布/审核通过时的自动生成共用本方法，保证静态化日志、
// 最后静态化时间（latest-times）与今日文件数口径一致。
// 去重键取自上游返回的 generated_at（每次生成唯一），避免同一结果被重复记录导致今日文件数虚高。
func (s *StaticLogService) RecordPageStaticDone(operator, operation, pageName string, statusCode int, body []byte) {
	if statusCode != http.StatusOK {
		return
	}
	var resp StaticPageResponse
	if err := json.Unmarshal(body, &resp); err != nil || !resp.OK || resp.Result == nil {
		return
	}
	res := resp.Result
	duration := "-"
	if res.DurationSeconds > 0 {
		duration = fmt.Sprintf("%.0f秒", res.DurationSeconds)
	}
	jobID := ""
	if res.GeneratedAt != "" {
		jobID = fmt.Sprintf("sync:%s:%s:%s", operation, pageName, res.GeneratedAt)
	}
	log := &models.StaticLog{
		Operation: operation,
		PageName:  pageName,
		Path:      res.Output,
		Duration:  duration,
		FileSize:  fmt.Sprintf("%d 个文件", res.GeneratedFiles),
		Operator:  operator,
		Status:    "success",
		Message:   fmt.Sprintf("页面：%s｜耗时：%.1f秒", pageName, res.DurationSeconds),
		JobID:     jobID,
	}
	if err := s.CreateIfNotExists(log); err != nil {
		utils.Logger.Warnf("记录静态化日志失败: %s", err)
	}
}

// recordDeleteLog 记录「删除静态文件」日志（详情页/专题页共用）。
// 去重键按 目标ID + 日期 生成，避免同一目标同一天内重复触发（如删除文章时多次调用）产生重复记录。
func (s *StaticLogService) recordDeleteLog(operator, operation, label, jobPrefix, targetID string, statusCode int, body []byte) {
	status, statusText := "danger", "删除失败"
	if statusCode == http.StatusOK {
		var resp struct {
			OK bool `json:"ok"`
		}
		if err := json.Unmarshal(body, &resp); err == nil && resp.OK {
			status, statusText = "success", "删除成功"
		}
	}
	log := &models.StaticLog{
		Operation: operation,
		PageName:  targetID,
		Path:      "-",
		Duration:  "-",
		FileSize:  "-",
		Operator:  operator,
		Status:    status,
		Message:   fmt.Sprintf("%s：%s｜%s", label, targetID, statusText),
		JobID:     fmt.Sprintf("%s:%s:%s", jobPrefix, targetID, time.Now().Format("2006-01-02")),
	}
	if err := s.CreateIfNotExists(log); err != nil {
		utils.Logger.Warnf("记录静态化日志失败: %s", err)
	}
}

// RecordArticleDeleteLog 记录「删除详情页静态文件」日志。
func (s *StaticLogService) RecordArticleDeleteLog(operator, articleID string, statusCode int, body []byte) {
	s.recordDeleteLog(operator, "删除详情页静态文件", "文章ID", "delete", articleID, statusCode, body)
}

// RecordTopicDeleteLog 记录「删除专题页静态文件」日志。
func (s *StaticLogService) RecordTopicDeleteLog(operator, topicID string, statusCode int, body []byte) {
	s.recordDeleteLog(operator, "删除专题页静态文件", "专题ID", "delete-topic", topicID, statusCode, body)
}

// latestSuccessTime 查询指定操作的最近一次成功静态化时间；operation 为空时表示不限操作（全站）
func (s *StaticLogService) latestSuccessTime(operation string) (string, error) {
	query := utils.DB.Model(&models.StaticLog{}).Where("status = ?", "success")
	if operation != "" {
		query = query.Where("operation LIKE ?", operation+"%")
	}
	var log models.StaticLog
	if err := query.Order("created_at DESC").First(&log).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "-", nil
		}
		return "", err
	}
	return log.CreatedAt.Format("2006-01-02 15:04:05"), nil
}

// GetLatestSuccessTimes 读取各类型页面最后一次静态化成功时间及今日成功生成文件数
func (s *StaticLogService) GetLatestSuccessTimes() (*StaticLatestTimes, error) {
	result := &StaticLatestTimes{}
	var err error
	if result.Site, err = s.latestSuccessTime("生成全站任务完成"); err != nil {
		return nil, err
	}
	if result.Home, err = s.latestSuccessTime("生成首页任务完成"); err != nil {
		return nil, err
	}
	if result.Column, err = s.latestSuccessTime("生成栏目页任务完成"); err != nil {
		return nil, err
	}
	if result.Topic, err = s.latestSuccessTime("生成专题页任务完成"); err != nil {
		return nil, err
	}
	if result.Detail, err = s.latestSuccessTime("生成详情页任务完成"); err != nil {
		return nil, err
	}
	if result.TodayFiles, err = s.GetTodayFileCount(); err != nil {
		return nil, err
	}
	return result, nil
}

// parseFileCount 从日志 file_size 文本（如 "65 个文件"）中解析文件数，无法解析时返回 0
func parseFileCount(s string) int {
	s = strings.TrimSpace(s)
	idx := strings.Index(s, "个文件")
	if idx <= 0 {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(s[:idx]))
	if err != nil {
		return 0
	}
	return n
}

// GetTodayFileCount 统计今日静态化成功生成的页面文件数（读取今日成功日志的 file_size 并求和）
func (s *StaticLogService) GetTodayFileCount() (int, error) {
	today := time.Now().Format("2006-01-02")
	var files []string
	if err := utils.DB.Model(&models.StaticLog{}).
		Where("status = ? AND DATE(created_at) = ?", "success", today).
		Pluck("file_size", &files).Error; err != nil {
		return 0, err
	}
	total := 0
	for _, f := range files {
		total += parseFileCount(f)
	}
	return total, nil
}
