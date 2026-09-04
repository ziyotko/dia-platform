package utils

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"server/config"
)

// RecordReplayFail 记录一次防重放/请求签名校验失败。
//
// 用途：
//   - 写入安全日志，便于事后审计与告警；
//   - 对明确的攻击特征（伪造签名、重复 nonce、畸形 nonce、过期时间戳）做计数；
//     达到 ReplayMaxFail 后临时封禁来源 IP，实现对持续探测/渗透的主动阻断。
//
// escalate 为 true 时才计入封禁阈值；缺失请求头等可能由客户端配置不当引起，
// 仅记日志、不触发封禁，避免误伤集成方。
func RecordReplayFail(c *gin.Context, reason string, userID uint, escalate bool) {
	ip := RealIP(c)

	Logger.WithFields(logrus.Fields{
		"event":     "replay_attack",
		"ip":        ip,
		"method":    c.Request.Method,
		"path":      c.Request.URL.Path,
		"reason":    reason,
		"userID":    userID,
		"timestamp": time.Now().UnixMilli(),
	}).Warn("防重放/请求签名校验失败")

	if !escalate {
		return
	}

	maxFail := config.AppConfig.Server.ReplayMaxFail
	banMinutes := config.AppConfig.Server.ReplayBanMinutes
	if maxFail <= 0 {
		maxFail = 10
	}
	if banMinutes <= 0 {
		banMinutes = 15
	}

	failKey := fmt.Sprintf("replay:fail:%s", ip)
	failCount, err := Redis1.Incr(Ctx, failKey).Result()
	if err != nil {
		return
	}
	if failCount == 1 {
		Redis1.Expire(Ctx, failKey, time.Duration(banMinutes)*time.Minute)
	}
	if failCount >= int64(maxFail) {
		Redis1.Set(Ctx, fmt.Sprintf("replay:ban:%s", ip), "1", time.Duration(banMinutes)*time.Minute)
	}
}

// GetReplayFailCount 返回某来源当前累计的失败次数（用于测试/诊断）。
func GetReplayFailCount(ip string) int64 {
	c, _ := Redis1.Get(Ctx, fmt.Sprintf("replay:fail:%s", ip)).Int64()
	return c
}
