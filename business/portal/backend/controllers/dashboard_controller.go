package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/services"
	"server/utils"
)

type DashboardController struct {
	articleService *services.ArticleService
	logService     *services.LogService
	userService    *services.UserService
}

func NewDashboardController() *DashboardController {
	return &DashboardController{
		articleService: &services.ArticleService{},
		logService:     &services.LogService{},
		userService:    &services.UserService{},
	}
}

func (c *DashboardController) GetStats(ctx *gin.Context) {
	articleCount := c.articleService.GetArticleCount()

	var todayVisit int64
	var myDraftCount int64
	var myArticleCount int64
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	utils.DB.Model(&models.VisitAnalytics{}).Where("visited_at >= ? AND visited_at < ?", startOfDay, endOfDay).Count(&todayVisit)

	userID := ctx.GetUint("userID")
	userIDStr := strconv.FormatUint(uint64(userID), 10)
	// 以下两项均为「我的」维度，与日期无关
	utils.DB.Model(&models.Article{}).Where("author_code = ? AND status = ?", userIDStr, models.ArticleStatusPublished).Count(&myArticleCount)
	utils.DB.Model(&models.Article{}).Where("author_code = ? AND status = ?", userIDStr, models.ArticleStatusDraft).Count(&myDraftCount)

	ctx.JSON(http.StatusOK, utils.Success("获取成功", gin.H{
		"articleCount":   articleCount,
		"todayVisit":     todayVisit,
		"myDraftCount":   myDraftCount,
		"myArticleCount": myArticleCount,
	}))
}

type LoginLogItem struct {
	Time     string `json:"time"`
	Username string `json:"username"`
	IP       string `json:"ip"`
	Browser  string `json:"browser"`
	OS       string `json:"os"`
	Device   string `json:"device"`
	Status   string `json:"status"`
}

type TrendItem struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}

// countVisitsByBucket 按 MySQL 时间格式分桶统计访问量（只回传分桶行，不在内存里逐条累加）。
// 原实现把整段时间范围内的记录全部读进内存再分桶，年窗口下会一次加载全年记录。
func countVisitsByBucket(mysqlLayout string, start, end time.Time) (map[string]int64, error) {
	var rows []struct {
		Bucket string `gorm:"column:bucket"`
		Total  int64  `gorm:"column:total"`
	}
	if err := utils.DB.Model(&models.VisitAnalytics{}).
		Select("DATE_FORMAT(visited_at, ?) AS bucket, COUNT(*) AS total", mysqlLayout).
		Where("visited_at >= ? AND visited_at < ?", start, end).
		Group("bucket").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	countMap := make(map[string]int64, len(rows))
	for _, row := range rows {
		countMap[row.Bucket] = row.Total
	}
	return countMap, nil
}

func (c *DashboardController) GetVisitTrend(ctx *gin.Context) {
	period := ctx.Query("period")
	now := time.Now()

	// 支持按年份查看全年趋势，默认当前年
	year := now.Year()
	if yStr := ctx.Query("year"); yStr != "" {
		if y, err := strconv.Atoi(yStr); err == nil && y >= 2000 && y <= 9999 {
			year = y
		}
	}

	var result []TrendItem

	switch period {
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		monday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(weekday - 1))
		nextMonday := monday.AddDate(0, 0, 7)

		countMap, err := countVisitsByBucket("%Y-%m-%d", monday, nextMonday)
		if err != nil {
			ctx.JSON(http.StatusOK, utils.Error(1, "获取访问趋势失败"))
			return
		}

		days := []string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"}
		for i := range 7 {
			day := monday.AddDate(0, 0, i)
			dateStr := day.Format("2006-01-02")
			result = append(result, TrendItem{Label: days[i], Value: countMap[dateStr]})
		}

	case "month":
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endOfMonth := startOfMonth.AddDate(0, 1, 0)
		daysInMonth := endOfMonth.AddDate(0, 0, -1).Day()

		countMap, err := countVisitsByBucket("%Y-%m-%d", startOfMonth, endOfMonth)
		if err != nil {
			ctx.JSON(http.StatusOK, utils.Error(1, "获取访问趋势失败"))
			return
		}

		for i := 1; i <= daysInMonth; i++ {
			day := time.Date(now.Year(), now.Month(), i, 0, 0, 0, 0, now.Location())
			dateStr := day.Format("2006-01-02")
			result = append(result, TrendItem{Label: day.Format("2日"), Value: countMap[dateStr]})
		}

	case "year":
		startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, now.Location())
		endOfYear := startOfYear.AddDate(1, 0, 0)

		countMap, err := countVisitsByBucket("%Y-%m", startOfYear, endOfYear)
		if err != nil {
			ctx.JSON(http.StatusOK, utils.Error(1, "获取访问趋势失败"))
			return
		}

		months := []string{"1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"}
		for i := 1; i <= 12; i++ {
			monthStr := time.Date(year, time.Month(i), 1, 0, 0, 0, 0, now.Location()).Format("2006-01")
			result = append(result, TrendItem{Label: months[i-1], Value: countMap[monthStr]})
		}

	default:
		ctx.JSON(http.StatusOK, utils.Error(1, "无效的 period 参数"))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("获取成功", result))
}

func (c *DashboardController) GetArticleTrend(ctx *gin.Context) {
	period := ctx.Query("period")
	now := time.Now()

	// 支持按年份查看全年趋势，默认当前年
	year := now.Year()
	if yStr := ctx.Query("year"); yStr != "" {
		if y, err := strconv.Atoi(yStr); err == nil && y >= 2000 && y <= 9999 {
			year = y
		}
	}

	var result []TrendItem

	switch period {
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		monday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(weekday - 1))
		nextMonday := monday.AddDate(0, 0, 7)

		countMap := getPublishedArticleCountMap(monday, nextMonday, "%Y-%m-%d")

		days := []string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"}
		for i := range 7 {
			day := monday.AddDate(0, 0, i)
			dateStr := day.Format("2006-01-02")
			result = append(result, TrendItem{Label: days[i], Value: countMap[dateStr]})
		}

	case "month":
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endOfMonth := startOfMonth.AddDate(0, 1, 0)
		daysInMonth := endOfMonth.AddDate(0, 0, -1).Day()

		countMap := getPublishedArticleCountMap(startOfMonth, endOfMonth, "%Y-%m-%d")

		for i := 1; i <= daysInMonth; i++ {
			day := time.Date(now.Year(), now.Month(), i, 0, 0, 0, 0, now.Location())
			dateStr := day.Format("2006-01-02")
			result = append(result, TrendItem{Label: day.Format("2日"), Value: countMap[dateStr]})
		}

	case "year":
		startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, now.Location())
		endOfYear := startOfYear.AddDate(1, 0, 0)

		countMap := getPublishedArticleCountMap(startOfYear, endOfYear, "%Y-%m")

		months := []string{"1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"}
		for i := 1; i <= 12; i++ {
			monthStr := time.Date(year, time.Month(i), 1, 0, 0, 0, 0, now.Location()).Format("2006-01")
			result = append(result, TrendItem{Label: months[i-1], Value: countMap[monthStr]})
		}

	default:
		ctx.JSON(http.StatusOK, utils.Error(1, "无效的 period 参数"))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("获取成功", result))
}

// getPublishedArticleCountMap 统计窗口内「发布」的文章数（按首次发布时间分桶）。
// 原先按 article.created_at 分桶，导致 3 月建稿、9 月发布的文章被计入 3 月（图表名为「发布趋势」）。
// 发布时间取 article_column_publish.created_at 的最小值（一篇文章首批多个栏目时可能存在多条发布记录）。
// mysqlLayout 为 MySQL 的 DATE_FORMAT 表达式（如 '%Y-%m'）。
func getPublishedArticleCountMap(start, end time.Time, mysqlLayout string) map[string]int64 {
	type bucketRow struct {
		Bucket string `gorm:"column:bucket"`
		Count  int64  `gorm:"column:cnt"`
	}
	var rows []bucketRow
	sub := utils.DB.Table("article_column_publish").
		Select("article_id, MIN(created_at) AS first_publish").
		Where("created_at >= ? AND created_at < ?", start, end).
		Group("article_id")
	if err := utils.DB.Table("(?) AS t", sub).
		Select(fmt.Sprintf("DATE_FORMAT(first_publish, '%s') AS bucket, COUNT(*) AS cnt", mysqlLayout)).
		Group("bucket").
		Scan(&rows).Error; err != nil {
		utils.Logger.Warnf("统计文章发布趋势失败: %v", err)
		return map[string]int64{}
	}
	countMap := make(map[string]int64, len(rows))
	for _, row := range rows {
		countMap[row.Bucket] = row.Count
	}
	return countMap
}

func (c *DashboardController) GetLoginLogs(ctx *gin.Context) {
	// 登录日志包含全站用户名的登录 IP/浏览器/操作系统，而「管理首页」菜单对内容审核/内容作者默认开放，
	// 因此非管理员只能看到自己的记录（原先无过滤地整站下发）。
	userID := ctx.GetUint("userID")
	var usernames []string
	if !models.HasAdminRoleIDs(c.userService.MustGetUserRoleIds(userID)) {
		if u, err := c.userService.GetUserByID(userID); err == nil && u != nil {
			for _, name := range []string{u.Username, u.Account, u.Email, u.Mobile} {
				if name != "" {
					usernames = append(usernames, name)
				}
			}
		}
		if len(usernames) == 0 {
			ctx.JSON(http.StatusOK, utils.Success("获取成功", []LoginLogItem{}))
			return
		}
	}

	logs, err := c.logService.GetRecentLoginLogs(10, userID, usernames)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取登录日志失败"))
		return
	}

	var list []LoginLogItem
	for _, log := range logs {
		status := "成功"
		if log.Status == 0 {
			status = "失败"
		}
		list = append(list, LoginLogItem{
			Time:     log.CreatedAt.Format("2006-01-02 15:04:05"),
			Username: log.Username,
			IP:       log.IP,
			Browser:  log.Browser,
			OS:       log.OS,
			Device:   log.Device,
			Status:   status,
		})
	}

	ctx.JSON(http.StatusOK, utils.Success("获取成功", list))
}
