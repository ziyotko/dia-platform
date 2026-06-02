package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"server/services"
	"server/utils"
)

type LogController struct {
	logService *services.LogService
}

func NewLogController() *LogController {
	return &LogController{
		logService: &services.LogService{},
	}
}

type LogListItem struct {
	ID          uint   `json:"id"`
	UserID      uint   `json:"userId"`
	Username    string `json:"username"`
	Type        string `json:"type"`
	Module      string `json:"module"`
	Description string `json:"description"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	Params      string `json:"params"`
	IP          string `json:"ip"`
	UserAgent   string `json:"ua"`
	Duration    int64  `json:"duration"`
	StatusCode  int    `json:"statusCode"`
	CreateTime  string `json:"createTime"`
}

func (c *LogController) GetLogs(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	username := ctx.Query("username")
	logType := ctx.Query("type")
	startDate := ctx.Query("startDate")
	endDate := ctx.Query("endDate")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	result, err := c.logService.GetLogList(page, pageSize, username, logType, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取日志列表失败"))
		return
	}

	list := make([]LogListItem, 0, len(result.List))
	for _, log := range result.List {
		list = append(list, LogListItem{
			ID:          log.ID,
			UserID:      log.UserID,
			Username:    log.Username,
			Type:        log.Type,
			Module:      log.Module,
			Description: log.Description,
			Method:      log.Method,
			Path:        log.Path,
			Params:      log.Params,
			IP:          log.IP,
			UserAgent:   log.UserAgent,
			Duration:    log.Duration,
			StatusCode:  log.StatusCode,
			CreateTime:  log.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	ctx.JSON(http.StatusOK, utils.Success("获取日志列表成功", gin.H{
		"list":  list,
		"total": result.Total,
	}))
}

func (c *LogController) ClearLogs(ctx *gin.Context) {
	err := c.logService.ClearLogs()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "清空日志失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("清空日志成功", nil))
}
