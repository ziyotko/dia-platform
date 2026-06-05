package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

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

	ctx.JSON(http.StatusOK, utils.Success("获取成功", gin.H{
		"adCount":      adCount,
		"articleCount": articleCount,
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
