package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"server/services"
	"server/utils"
)

type StaticMonitorController struct {
	settingsService *services.SettingsService
}

func NewStaticMonitorController() *StaticMonitorController {
	return &StaticMonitorController{
		settingsService: &services.SettingsService{},
	}
}

// GetStaticMonitor 静态化服务监控：后台调用静态化程序访问地址全局变量 + /healthz 检查服务运行状态
// 成功条件：HTTP 状态码 200 且响应体为 {"ok": true}
func (c *StaticMonitorController) GetStaticMonitor(ctx *gin.Context) {
	checkTime := time.Now().Format("2006-01-02 15:04:05")

	params, err := c.settingsService.GetStaticParamsFromCache()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取静态化参数失败"))
		return
	}

	addr := strings.TrimSpace(params.StaticProgramAddr)
	if addr == "" {
		ctx.JSON(http.StatusOK, utils.Success("静态化服务监控", gin.H{
			"online":        false,
			"address":       "",
			"httpStatus":    0,
			"lastCheckTime": checkTime,
			"message":       "静态化程序访问地址未配置",
		}))
		return
	}

	// 规范化访问地址：缺少协议时默认 http；显式写了非 http/https 协议视为配置错误
	// （与 services.staticProgramBaseURL 的口径保持一致，否则「监控显示在线」而实际调用全失败）
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		if strings.Contains(addr, "://") {
			ctx.JSON(http.StatusOK, utils.Success("静态化服务监控", gin.H{
				"online":        false,
				"address":       addr,
				"httpStatus":    0,
				"lastCheckTime": checkTime,
				"message":       "静态化程序访问地址只支持 http:// 或 https:// 开头，请在「系统设置-静态化设置」中修正",
			}))
			return
		}
		addr = "http://" + addr
	}
	healthURL := strings.TrimRight(addr, "/") + "/healthz"

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(healthURL)
	if err != nil {
		utils.Logger.Warnf("静态化服务健康检查失败: %s, url=%s", err, healthURL)
		ctx.JSON(http.StatusOK, utils.Success("静态化服务监控", gin.H{
			"online":        false,
			"address":       healthURL,
			"httpStatus":    0,
			"lastCheckTime": checkTime,
			"message":       "静态化服务未运行或无法连接",
		}))
		return
	}
	defer resp.Body.Close()

	// 成功条件：HTTP 200 且响应体为 {"ok": true}；
	// 仅判状态码会把上游返回 200 + {"ok": false} 误报为“运行正常”，进而放行高危操作。
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	healthy := false
	if resp.StatusCode == http.StatusOK {
		var payload struct {
			OK bool `json:"ok"`
		}
		if err := json.Unmarshal(body, &payload); err == nil && payload.OK {
			healthy = true
		}
	}

	if healthy {
		ctx.JSON(http.StatusOK, utils.Success("静态化服务监控", gin.H{
			"online":        true,
			"address":       healthURL,
			"httpStatus":    resp.StatusCode,
			"lastCheckTime": checkTime,
			"message":       "静态化服务运行正常",
		}))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("静态化服务监控", gin.H{
		"online":        false,
		"address":       healthURL,
		"httpStatus":    resp.StatusCode,
		"lastCheckTime": checkTime,
		"message":       "静态化服务响应异常（HTTP 状态码: " + strconv.Itoa(resp.StatusCode) + "，期望 200 且 {\"ok\": true}）",
	}))
}
