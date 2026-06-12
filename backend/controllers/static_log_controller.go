package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"server/services"
	"server/utils"
)

type StaticLogController struct {
	staticLogService *services.StaticLogService
}

func NewStaticLogController() *StaticLogController {
	return &StaticLogController{
		staticLogService: &services.StaticLogService{},
	}
}

type StaticLogListItem struct {
	ID        uint   `json:"id"`
	Operation string `json:"operation"`
	PageName  string `json:"pageName"`
	Path      string `json:"path"`
	Duration  string `json:"duration"`
	FileSize  string `json:"fileSize"`
	Operator  string `json:"operator"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	CreateTime string `json:"createTime"`
}

func (c *StaticLogController) GetLogs(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	result, err := c.staticLogService.GetList(page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取静态化日志列表失败"))
		return
	}

	list := make([]StaticLogListItem, 0, len(result.List))
	for _, log := range result.List {
		list = append(list, StaticLogListItem{
			ID:         log.ID,
			Operation:  log.Operation,
			PageName:   log.PageName,
			Path:       log.Path,
			Duration:   log.Duration,
			FileSize:   log.FileSize,
			Operator:   log.Operator,
			Status:     log.Status,
			Message:    log.Message,
			CreateTime: log.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	ctx.JSON(http.StatusOK, utils.Success("获取静态化日志列表成功", gin.H{
		"list":  list,
		"total": result.Total,
	}))
}

func (c *StaticLogController) ClearLogs(ctx *gin.Context) {
	err := c.staticLogService.Clear()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "清空静态化日志失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("清空静态化日志成功", nil))
}
