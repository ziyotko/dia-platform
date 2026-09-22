package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/utils"
)

// AnalyticsController 文章数据分析统计（点赞/分享/浏览）
type AnalyticsController struct{}

func NewAnalyticsController() *AnalyticsController {
	return &AnalyticsController{}
}

// analyticsWindow 统计时间窗口 [Start, End)
type analyticsWindow struct {
	Start time.Time
	End   time.Time
}

// GetArticleAnalyticsTrend 获取文章点赞/分享/浏览趋势
// period: week(本周按天) / month(本月按天) / year(全年按月)
func (c *AnalyticsController) GetArticleAnalyticsTrend(ctx *gin.Context) {
	period := ctx.Query("period")
	if period == "" {
		period = "week"
	}

	now := time.Now()
	loc := now.Location()

	// period=year 时支持按年份查看全年趋势（与仪表盘趋势接口一致），默认当前年
	year := now.Year()
	if yStr := ctx.Query("year"); yStr != "" {
		if y, err := strconv.Atoi(yStr); err == nil && y >= 2000 && y <= 9999 {
			year = y
		}
	}

	labels, windows, err := buildAnalyticsWindows(period, year, now, loc)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "无效的 period 参数"))
		return
	}

	// 桶粒度与窗口边界对齐：week/month 按天、year 按月（与 buildAnalyticsWindows 一致）
	mysqlLayout, goLayout := "%Y-%m-%d", "2006-01-02"
	if period == "year" {
		mysqlLayout, goLayout = "%Y-%m", "2006-01"
	}
	like := countAnalyticsSeries(&models.LikeAnalytics{}, "liked_at", mysqlLayout, goLayout, windows)
	share := countAnalyticsSeries(&models.ShareAnalytics{}, "shared_at", mysqlLayout, goLayout, windows)
	visit := countAnalyticsSeries(&models.VisitAnalytics{}, "visited_at", mysqlLayout, goLayout, windows)

	ctx.JSON(http.StatusOK, utils.Success("获取成功", gin.H{
		"labels":     labels,
		"like":       like,
		"share":      share,
		"visit":      visit,
		"totalLike":  sumInt64(like),
		"totalShare": sumInt64(share),
		"totalVisit": sumInt64(visit),
	}))
}

// buildAnalyticsWindows 根据周期构建时间窗口与对应标签（year 仅对 period=year 生效）
func buildAnalyticsWindows(period string, year int, now time.Time, loc *time.Location) ([]string, []analyticsWindow, error) {
	switch period {
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		monday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, -(weekday - 1))
		labels := []string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"}
		windows := make([]analyticsWindow, 7)
		for i := 0; i < 7; i++ {
			day := monday.AddDate(0, 0, i)
			windows[i] = analyticsWindow{Start: day, End: day.AddDate(0, 0, 1)}
		}
		return labels, windows, nil
	case "month":
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		daysInMonth := startOfMonth.AddDate(0, 1, -1).Day()
		labels := make([]string, 0, daysInMonth)
		windows := make([]analyticsWindow, 0, daysInMonth)
		for i := 1; i <= daysInMonth; i++ {
			day := time.Date(now.Year(), now.Month(), i, 0, 0, 0, 0, loc)
			labels = append(labels, day.Format("2日"))
			windows = append(windows, analyticsWindow{Start: day, End: day.AddDate(0, 0, 1)})
		}
		return labels, windows, nil
	case "year":
		labels := []string{"1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"}
		windows := make([]analyticsWindow, 12)
		for i := 1; i <= 12; i++ {
			start := time.Date(year, time.Month(i), 1, 0, 0, 0, 0, loc)
			windows[i-1] = analyticsWindow{Start: start, End: start.AddDate(0, 1, 0)}
		}
		return labels, windows, nil
	default:
		return nil, nil, fmt.Errorf("无效的 period 参数: %s", period)
	}
}

// countAnalyticsSeries 统计某一 analytics 表在各时间窗口内的记录数。
// 统计在 SQL 里按桶（天/月）完成，只回传桶行；原实现把整段范围内的记录全部读进内存再逐条分桶。
func countAnalyticsSeries(model any, timeCol, mysqlLayout, goLayout string, windows []analyticsWindow) []int64 {
	values := make([]int64, len(windows))
	if len(windows) == 0 {
		return values
	}

	// 窗口与桶边界对齐（按天或按月），因此可以直接用「桶 → 窗口」映射
	index := make(map[string]int, len(windows))
	for i := range windows {
		index[windows[i].Start.Format(goLayout)] = i
	}

	var rows []struct {
		Bucket string `gorm:"column:bucket"`
		Total  int64  `gorm:"column:total"`
	}
	start := windows[0].Start
	end := windows[len(windows)-1].End
	if err := utils.DB.Model(model).
		Select("DATE_FORMAT("+timeCol+", ?) AS bucket, COUNT(*) AS total", mysqlLayout).
		Where(timeCol+" >= ? AND "+timeCol+" < ?", start, end).
		Group("bucket").
		Scan(&rows).Error; err != nil {
		utils.Logger.Warnf("统计 %s 趋势失败: %s", timeCol, err)
		return values
	}
	for _, row := range rows {
		if i, ok := index[row.Bucket]; ok {
			values[i] = row.Total
		}
	}
	return values
}

func sumInt64(values []int64) int64 {
	var total int64
	for _, v := range values {
		total += v
	}
	return total
}
