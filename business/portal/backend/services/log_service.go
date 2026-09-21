package services

import (
	"strings"
	"time"

	"server/models"
	"server/utils"

	"gorm.io/gorm"
)

// parseLogDate 解析日志筛选的日期参数（YYYY-MM-DD）。
// 日期筛选两端均可单独使用；原实现要求 startDate 与 endDate 同时提供，只传一端会被静默忽略。
func parseLogDate(value string) (time.Time, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation("2006-01-02", trimmed, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// applyLogDateRange 把日期区间条件加到查询上。
// 用 created_at >= / < 的范围条件（半开区间），而不是 DATE(created_at) BETWEEN：
// 后者会让 created_at 上的索引失效，且无法正确处理“只传一端”。
func applyLogDateRange(query *gorm.DB, startDate, endDate string) *gorm.DB {
	if start, ok := parseLogDate(startDate); ok {
		query = query.Where("created_at >= ?", start)
	}
	if end, ok := parseLogDate(endDate); ok {
		query = query.Where("created_at < ?", end.AddDate(0, 0, 1)) // 含当天
	}
	return query
}

type LogService struct{}

type LogListResult struct {
	Total int64                 `json:"total"`
	List  []models.OperationLog `json:"list"`
}

func (s *LogService) GetLogList(page, pageSize int, username, logType string, startDate, endDate string) (*LogListResult, error) {
	var logs []models.OperationLog
	var total int64

	query := utils.DB.Model(&models.OperationLog{})

	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if logType != "" {
		query = query.Where("type = ?", logType)
	}
	query = applyLogDateRange(query, startDate, endDate)

	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	err = query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	if err != nil {
		return nil, err
	}

	return &LogListResult{
		Total: total,
		List:  logs,
	}, nil
}

// ClearLogs 仅清空半年前的日志，系统必须保留最近半年的日志。返回删除条数。
func (s *LogService) ClearLogs() (int64, error) {
	cutoff := time.Now().AddDate(0, -6, 0)
	result := utils.DB.Where("created_at < ?", cutoff).Delete(&models.OperationLog{})
	return result.RowsAffected, result.Error
}

func (s *LogService) CreateLog(log *models.OperationLog) error {
	return utils.DB.Create(log).Error
}

func (s *LogService) GetUserOperationCount(userID uint) (int64, error) {
	var count int64
	err := utils.DB.Model(&models.OperationLog{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

type LoginLogListResult struct {
	Total int64             `json:"total"`
	List  []models.LoginLog `json:"list"`
}

func (s *LogService) GetLoginLogList(page, pageSize int, username, status, startDate, endDate string) (*LoginLogListResult, error) {
	var logs []models.LoginLog
	var total int64

	query := utils.DB.Model(&models.LoginLog{})

	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query = applyLogDateRange(query, startDate, endDate)

	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	err = query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	if err != nil {
		return nil, err
	}

	return &LoginLogListResult{
		Total: total,
		List:  logs,
	}, nil
}

// GetRecentLoginLogs 最近登录日志。
// usernames 非空时只返回这些用户名的记录（登录成功记录显示名、失败记录登录输入值，故调用方传多个候选）；
// userID 非 0 时同时限定 user_id，避免同名账号（username 非唯一）的登录 IP/浏览器泄露给他人。
// 历史行 user_id 为 0，仍按用户名匹配（否则升级后本人看不到旧记录）；新行按 user_id 精确过滤。
func (s *LogService) GetRecentLoginLogs(limit int, userID uint, usernames []string) ([]models.LoginLog, error) {
	var logs []models.LoginLog
	query := utils.DB.Model(&models.LoginLog{})
	if len(usernames) > 0 {
		if userID > 0 {
			query = query.Where("username IN ? AND (user_id = 0 OR user_id = ?)", usernames, userID)
		} else {
			query = query.Where("username IN ?", usernames)
		}
	}
	err := query.Order("id DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

// ClearLoginLogs 与 ClearLogs 保持同一语义：仅清空半年前的登录日志，保留最近半年。返回删除条数。
func (s *LogService) ClearLoginLogs() (int64, error) {
	cutoff := time.Now().AddDate(0, -6, 0)
	result := utils.DB.Where("created_at < ?", cutoff).Delete(&models.LoginLog{})
	return result.RowsAffected, result.Error
}
