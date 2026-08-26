package services

import (
	"errors"
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

// StaticLatestTimes 各类型页面最后一次静态化成功时间（读取静态化日志最新成功记录）
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

func (s *StaticLogService) Clear() error {
	return utils.DB.Where("1 = 1").Unscoped().Delete(&models.StaticLog{}).Error
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
