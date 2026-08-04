package controllers

import (
	"fmt"
	"net/http"
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

	labels, windows, err := buildAnalyticsWindows(period, now, loc)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "无效的 period 参数"))
		return
	}

	like := countAnalyticsSeries(&models.LikeAnalytics{}, "liked_at", windows)
	share := countAnalyticsSeries(&models.ShareAnalytics{}, "shared_at", windows)
	visit := countAnalyticsSeries(&models.VisitAnalytics{}, "visited_at", windows)

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

// buildAnalyticsWindows 根据周期构建时间窗口与对应标签
func buildAnalyticsWindows(period string, now time.Time, loc *time.Location) ([]string, []analyticsWindow, error) {
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
			start := time.Date(now.Year(), time.Month(i), 1, 0, 0, 0, 0, loc)
			windows[i-1] = analyticsWindow{Start: start, End: start.AddDate(0, 1, 0)}
		}
		return labels, windows, nil
	default:
		return nil, nil, fmt.Errorf("无效的 period 参数: %s", period)
	}
}

// countAnalyticsSeries 统计某一 analytics 表在各时间窗口内的记录数（单次查询，内存分桶）
func countAnalyticsSeries(model any, timeCol string, windows []analyticsWindow) []int64 {
	values := make([]int64, len(windows))
	if len(windows) == 0 {
		return values
	}

	indexByDate := make(map[string]int, len(windows))
	for i, w := range windows {
		indexByDate[w.Start.Format("2006-01-02")] = i
	}

	var rows []struct {
		At time.Time `gorm:"column:at"`
	}
	start := windows[0].Start
	end := windows[len(windows)-1].End
	utils.DB.Model(model).
		Select(timeCol+" AS at").
		Where(timeCol+" >= ? AND "+timeCol+" < ?", start, end).
		Scan(&rows)

	for _, row := range rows {
		if idx, ok := indexByDate[row.At.Format("2006-01-02")]; ok {
			values[idx]++
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
