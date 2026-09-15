package service

import (
	"time"

	"base/internal/models"
	"base/pkg/db"
)

type DashboardService struct{}

// DailyPoint 趋势图上的一个点（按天聚合）。
type DailyPoint struct {
	Date  string `json:"date"` // YYYY-MM-DD
	Count int64  `json:"count"`
}

// StatsResult 仪表盘统计。
//
// 口径总览：
//   - 平台超管：全平台维度；
//   - 租户管理员：本租户维度（机构/消息等「平台预置 + 租户自建」的数据含 tenant_id = 0 的共享部分）；
//   - 普通用户：资源汇总类指标不返回（保持 0），只给「已开通应用」与「我的」数据；
//   - 「我的」类指标（未读、待办、我发起的流程）始终只看当前用户；
//   - 趋势为最近 7 天（含今天），没有数据的日期补 0。
type StatsResult struct {
	// —— 资源概览 ——
	TenantCount        int64 `json:"tenantCount"`        // 平台：租户总数；租户用户：0
	TenantEnabledCount int64 `json:"tenantEnabledCount"` // 平台：启用中的租户数
	AppCount           int64 `json:"appCount"`           // 平台：应用总数；租户：已开通且启用的应用数
	AppIframeCount     int64 `json:"appIframeCount"`     // 平台：iframe 接入的应用数
	AppProxyCount      int64 `json:"appProxyCount"`      // 平台：API 代理接入的应用数
	AppInstanceCount   int64 `json:"appInstanceCount"`   // 平台：应用实例总数；租户：本租户开通数
	UserCount          int64 `json:"userCount"`          // 上限：平台/租户管理员的统计范围
	UserEnabledCount   int64 `json:"userEnabledCount"`
	UserDisabledCount  int64 `json:"userDisabledCount"`
	UserAdminCount     int64 `json:"userAdminCount"`
	RoleCount          int64 `json:"roleCount"`
	OrganizationCount  int64 `json:"organizationCount"`
	MessageCount       int64 `json:"messageCount"`

	// —— 我的待办（当前用户）——
	MyUnreadCount          int64 `json:"myUnreadCount"`          // 未读站内信
	MyTodoCount            int64 `json:"myTodoCount"`            // 待我审批的流程任务
	MyRunningInstanceCount int64 `json:"myRunningInstanceCount"` // 我发起的、审批中的流程

	// —— 运行态 ——
	RunningInstanceCount int64 `json:"runningInstanceCount"` // 审批中的实例（平台：全部；租户管理员：本租户；普通用户：与我相关）

	// —— 今日 ——
	TodayNewUserCount   int64 `json:"todayNewUserCount"`
	TodayLoginCount     int64 `json:"todayLoginCount"`     // 今日登录成功次数
	TodayLoginFailCount int64 `json:"todayLoginFailCount"` // 今日登录失败次数

	// —— 近 7 天趋势 ——
	UserTrend    []DailyPoint `json:"userTrend"`
	LoginTrend   []DailyPoint `json:"loginTrend"`
	MessageTrend []DailyPoint `json:"messageTrend"`
}

// trendDays 趋势统计的天数（含今天）。
const trendDays = 7

// GetStats 汇总仪表盘统计。
// tenantID/userID 来自 JWT 上下文，isAdmin 为 base_user.is_admin。
func (s DashboardService) GetStats(tenantID, userID uint64, isAdmin bool) (*StatsResult, error) {
	res := &StatsResult{}
	platform := models.IsPlatformTenant(tenantID)
	// tenantWide：能否查看租户/平台维度的资源汇总（普通用户只能看「我的」）
	tenantWide := platform || isAdmin

	if platform {
		if err := countRows(&models.Tenant{}, "", nil, &res.TenantCount); err != nil {
			return nil, err
		}
		if err := countRows(&models.Tenant{}, "status = ?", []interface{}{1}, &res.TenantEnabledCount); err != nil {
			return nil, err
		}
		if err := countRows(&models.App{}, "", nil, &res.AppCount); err != nil {
			return nil, err
		}
		if err := countRows(&models.App{}, "type = ?", []interface{}{"iframe"}, &res.AppIframeCount); err != nil {
			return nil, err
		}
		if err := countRows(&models.App{}, "type = ?", []interface{}{"proxy"}, &res.AppProxyCount); err != nil {
			return nil, err
		}
		if err := countRows(&models.AppInstance{}, "", nil, &res.AppInstanceCount); err != nil {
			return nil, err
		}
	} else {
		// 租户：已开通且启用的应用数 + 本租户开通的实例数
		if err := db.DB.Model(&models.App{}).
			Joins("JOIN base_app_instance ON base_app_instance.app_id = base_app.id").
			Where("base_app_instance.tenant_id = ? AND base_app_instance.status = ? AND base_app.status = ?", tenantID, 1, 1).
			Count(&res.AppCount).Error; err != nil {
			return nil, err
		}
		if err := countRows(&models.AppInstance{}, "tenant_id = ?", []interface{}{tenantID}, &res.AppInstanceCount); err != nil {
			return nil, err
		}
	}

	today := time.Now().Format("2006-01-02")
	days := recentDays(trendDays)

	if tenantWide {
		userScope, userArgs := tenantOnlyScope(tenantID)
		if err := countRows(&models.User{}, userScope, userArgs, &res.UserCount); err != nil {
			return nil, err
		}
		if err := countRows(&models.User{}, andCond(userScope, "status = ?"), append(userArgs, 1), &res.UserEnabledCount); err != nil {
			return nil, err
		}
		if err := countRows(&models.User{}, andCond(userScope, "status <> ?"), append(userArgs, 1), &res.UserDisabledCount); err != nil {
			return nil, err
		}
		if err := countRows(&models.User{}, andCond(userScope, "is_admin = ?"), append(userArgs, true), &res.UserAdminCount); err != nil {
			return nil, err
		}
		if err := countRows(&models.Role{}, userScope, userArgs, &res.RoleCount); err != nil {
			return nil, err
		}

		sharedScope, sharedArgs := tenantSharedScope(tenantID)
		if err := countRows(&models.Organization{}, sharedScope, sharedArgs, &res.OrganizationCount); err != nil {
			return nil, err
		}
		if err := countRows(&models.Message{}, sharedScope, sharedArgs, &res.MessageCount); err != nil {
			return nil, err
		}

		loginScope, loginArgs := tenantOnlyScope(tenantID)
		if err := countRows(&models.User{}, andCond(userScope, "DATE(created_at) = ?"), append(userArgs, today), &res.TodayNewUserCount); err != nil {
			return nil, err
		}
		if err := countRows(&models.LoginLog{}, andCond(loginScope, "DATE(created_at) = ? AND status = ?"), append(loginArgs, today, 1), &res.TodayLoginCount); err != nil {
			return nil, err
		}
		if err := countRows(&models.LoginLog{}, andCond(loginScope, "DATE(created_at) = ? AND status = ?"), append(loginArgs, today, 0), &res.TodayLoginFailCount); err != nil {
			return nil, err
		}

		var err error
		if res.UserTrend, err = dailyCount(&models.User{}, userScope, userArgs, days); err != nil {
			return nil, err
		}
		if res.LoginTrend, err = dailyCount(&models.LoginLog{}, andCond(loginScope, "status = ?"), append(loginArgs, 1), days); err != nil {
			return nil, err
		}
		if res.MessageTrend, err = dailyCount(&models.Message{}, sharedScope, sharedArgs, days); err != nil {
			return nil, err
		}
	}

	// —— 我的待办（所有用户都看自己的）——
	unread, err := (MessageService{}).GetUnreadCount(userID, tenantID)
	if err != nil {
		return nil, err
	}
	res.MyUnreadCount = unread

	todoQuery := db.DB.Model(&models.WorkflowTask{}).
		Where("approver_id = ? AND status = ?", userID, models.WorkflowTaskPending)
	if !platform {
		todoQuery = todoQuery.Where("tenant_id = ?", tenantID)
	}
	if err := todoQuery.Count(&res.MyTodoCount).Error; err != nil {
		return nil, err
	}

	mineQuery := db.DB.Model(&models.WorkflowInstance{}).
		Where("initiator_id = ? AND status = ?", userID, models.WorkflowInstanceRunning)
	if !platform {
		mineQuery = mineQuery.Where("tenant_id = ?", tenantID)
	}
	if err := mineQuery.Count(&res.MyRunningInstanceCount).Error; err != nil {
		return nil, err
	}

	// —— 审批中的实例 ——
	runningQuery := db.DB.Model(&models.WorkflowInstance{}).Where("status = ?", models.WorkflowInstanceRunning)
	switch {
	case platform:
		// 全平台
	case isAdmin:
		runningQuery = runningQuery.Where("tenant_id = ?", tenantID)
	default:
		// 普通用户：我发起的 + 待我审批的（子查询去重）
		pending := db.DB.Model(&models.WorkflowTask{}).
			Select("instance_id").
			Where("approver_id = ? AND status = ?", userID, models.WorkflowTaskPending)
		runningQuery = runningQuery.
			Where("tenant_id = ? AND (initiator_id = ? OR id IN (?))", tenantID, userID, pending)
	}
	if err := runningQuery.Count(&res.RunningInstanceCount).Error; err != nil {
		return nil, err
	}

	return res, nil
}

// countRows 按条件统计行数（scope 为空表示不加条件）。
func countRows(model interface{}, scope string, args []interface{}, out *int64) error {
	query := db.DB.Model(model)
	if scope != "" {
		query = query.Where(scope, args...)
	}
	return query.Count(out).Error
}

// tenantOnlyScope 严格租户范围：平台超管不加条件。
func tenantOnlyScope(tenantID uint64) (string, []interface{}) {
	if models.IsPlatformTenant(tenantID) {
		return "", nil
	}
	return "tenant_id = ?", []interface{}{tenantID}
}

// tenantSharedScope 「本租户 + 平台内置（tenant_id = 0）」范围：平台超管不加条件。
func tenantSharedScope(tenantID uint64) (string, []interface{}) {
	if models.IsPlatformTenant(tenantID) {
		return "", nil
	}
	return "tenant_id = ? OR tenant_id = 0", []interface{}{tenantID}
}

func andCond(scope, extra string) string {
	if scope == "" {
		return extra
	}
	return scope + " AND " + extra
}

// recentDays 返回最近 n 天（含今天）的日期列表，升序。
func recentDays(n int) []string {
	days := make([]string, 0, n)
	today := time.Now()
	for i := n - 1; i >= 0; i-- {
		days = append(days, today.AddDate(0, 0, -i).Format("2006-01-02"))
	}
	return days
}

// dailyCount 按天聚合统计，缺少数据的日期补 0。
func dailyCount(model interface{}, scope string, args []interface{}, days []string) ([]DailyPoint, error) {
	type row struct {
		Day   string
		Total int64
	}
	query := db.DB.Model(model).
		Select("DATE(created_at) AS day, COUNT(*) AS total").
		Where("DATE(created_at) >= ?", days[0])
	if scope != "" {
		query = query.Where(scope, args...)
	}
	var rows []row
	if err := query.Group("DATE(created_at)").Scan(&rows).Error; err != nil {
		return nil, err
	}

	counts := make(map[string]int64, len(rows))
	for _, r := range rows {
		counts[normalizeDay(r.Day)] = r.Total
	}
	points := make([]DailyPoint, 0, len(days))
	for _, d := range days {
		points = append(points, DailyPoint{Date: d, Count: counts[d]})
	}
	return points, nil
}

// normalizeDay 把驱动返回的日期统一成 YYYY-MM-DD（不同驱动/时区下可能是 RFC3339）。
func normalizeDay(raw string) string {
	if len(raw) >= 10 {
		return raw[:10]
	}
	return raw
}
