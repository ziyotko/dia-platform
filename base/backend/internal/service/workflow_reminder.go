package service

import (
	"time"

	"base/config"

	"github.com/sirupsen/logrus"
)

// StartWorkflowReminder 启动工作流超时提醒的后台轮询。
// 扫描间隔由 server.workflow_remind_interval_seconds 配置，<= 0 表示关闭。
// 由 main 以 goroutine 方式启动（进程退出即结束，无需额外关闭逻辑）。
func StartWorkflowReminder() {
	interval := config.Cfg.Server.WorkflowRemindInterval
	if interval <= 0 {
		logrus.Info("工作流超时提醒已关闭（workflow_remind_interval_seconds <= 0）")
		return
	}
	logrus.Infof("工作流超时提醒已启动，扫描间隔 %d 秒", interval)

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		count, err := (WorkflowEngineService{}).RemindOverdueTasks()
		if err != nil {
			logrus.WithError(err).Warn("工作流超时提醒扫描失败")
			continue
		}
		if count > 0 {
			logrus.Infof("工作流超时提醒：本轮催办 %d 个待办", count)
		}
	}
}
