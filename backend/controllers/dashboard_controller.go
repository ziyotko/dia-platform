package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/services"
	"server/utils"
)

type DashboardController struct {
	adService      *services.AdService
	articleService *services.ArticleService
	logService     *services.LogService
}

func NewDashboardController() *DashboardController {
	return &DashboardController{
		adService:      &services.AdService{},
		articleService: &services.ArticleService{},
		logService:     &services.LogService{},
	}
}

func (c *DashboardController) GetStats(ctx *gin.Context) {
	adCount := c.adService.GetAdCount()
	articleCount := c.articleService.GetArticleCount()

	var todayVisit int64
	var todayStaticCount int64
	var todayAuditCount int64
	var myArticleCount int64
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	utils.DB.Model(&models.SiteAnalytics{}).Where("visited_at >= ? AND visited_at < ?", startOfDay, endOfDay).Count(&todayVisit)
	utils.DB.Model(&models.Article{}).Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).Count(&todayStaticCount)
	utils.DB.Model(&models.Article{}).Where("created_at >= ? AND created_at < ? AND audit_status = ?", startOfDay, endOfDay, 0).Count(&todayAuditCount)

	userID := ctx.GetUint("userID")
	utils.DB.Model(&models.Article{}).Where("author_code = ?", strconv.FormatUint(uint64(userID), 10)).Count(&myArticleCount)

	ctx.JSON(http.StatusOK, utils.Success("获取成功", gin.H{
		"adCount":          adCount,
		"articleCount":     articleCount,
		"todayVisit":       todayVisit,
		"todayStaticCount": todayStaticCount,
		"todayAuditCount":  todayAuditCount,
		"myArticleCount":   myArticleCount,
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

func (c *DashboardController) GetVisitTrend(ctx *gin.Context) {
	period := ctx.Query("period")
	now := time.Now()

	var result []TrendItem

	switch period {
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		monday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(weekday - 1))
		nextMonday := monday.AddDate(0, 0, 7)

		var visits []models.SiteAnalytics
		utils.DB.Where("visited_at >= ? AND visited_at < ?", monday, nextMonday).Find(&visits)

		countMap := make(map[string]int64)
		for _, v := range visits {
			dateStr := v.VisitedAt.Format("2006-01-02")
			countMap[dateStr]++
		}

		days := []string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"}
		for i := 0; i < 7; i++ {
			day := monday.AddDate(0, 0, i)
			dateStr := day.Format("2006-01-02")
			result = append(result, TrendItem{Label: days[i], Value: countMap[dateStr]})
		}

	case "month":
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endOfMonth := startOfMonth.AddDate(0, 1, 0)
		daysInMonth := endOfMonth.AddDate(0, 0, -1).Day()

		var visits []models.SiteAnalytics
		utils.DB.Where("visited_at >= ? AND visited_at < ?", startOfMonth, endOfMonth).Find(&visits)

		countMap := make(map[string]int64)
		for _, v := range visits {
			dateStr := v.VisitedAt.Format("2006-01-02")
			countMap[dateStr]++
		}

		for i := 1; i <= daysInMonth; i++ {
			day := time.Date(now.Year(), now.Month(), i, 0, 0, 0, 0, now.Location())
			dateStr := day.Format("2006-01-02")
			result = append(result, TrendItem{Label: day.Format("2日"), Value: countMap[dateStr]})
		}

	case "year":
		startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		endOfYear := startOfYear.AddDate(1, 0, 0)

		var visits []models.SiteAnalytics
		utils.DB.Where("visited_at >= ? AND visited_at < ?", startOfYear, endOfYear).Find(&visits)

		countMap := make(map[string]int64)
		for _, v := range visits {
			monthStr := v.VisitedAt.Format("2006-01")
			countMap[monthStr]++
		}

		months := []string{"1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"}
		for i := 1; i <= 12; i++ {
			monthStr := time.Date(now.Year(), time.Month(i), 1, 0, 0, 0, 0, now.Location()).Format("2006-01")
			result = append(result, TrendItem{Label: months[i-1], Value: countMap[monthStr]})
		}

	default:
		ctx.JSON(http.StatusOK, utils.Error(1, "无效的 period 参数"))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("获取成功", result))
}

func (c *DashboardController) GetLoginLogs(ctx *gin.Context) {
	logs, err := c.logService.GetRecentLoginLogs(5)
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
