package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"server/config"
	"server/utils"
)

// ReplayProtectionMiddleware 防重放 + 请求签名校验。
//
// 防护要点：
//  1. 时间戳新鲜度校验（默认窗口 120s，可从配置调整）；
//  2. 随机数（nonce）一次性使用，且绑定用户，避免跨用户碰撞与跨会话复用；
//  3. 已认证接口强制校验 HMAC-SHA256 请求签名，绑定方法/路径/时间戳/随机数/请求体哈希；
//  4. 主动检测：对伪造签名、重复 nonce、畸形/过期时间戳等明确攻击特征进行计数，
//     达到阈值后临时封禁来源 IP，并写入安全日志，便于审计与告警。
//
// 注意：公开（未认证）接口无法校验签名，仅能依赖时间戳 + nonce + 限流；
// 对 /visit、/like、/share 等公开写接口，请务必在业务层配合幂等去重与 IP 限流。
func ReplayProtectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		replayWindow := time.Duration(config.AppConfig.Server.ReplayWindowSecs) * time.Second
		if replayWindow <= 0 {
			replayWindow = 2 * time.Minute
		}

		timestampStr := c.GetHeader("X-Request-Timestamp")
		nonce := c.GetHeader("X-Request-Nonce")
		signature := c.GetHeader("X-Request-Signature")

		ip := utils.RealIP(c)
		userID, hasAuth := c.Get("userID")
		var uid uint
		if hasAuth {
			uid, _ = userID.(uint)
		}

		// 0. 来源已被临时封禁（持续重放/伪造签名）
		if utils.Redis1.Exists(utils.Ctx, fmt.Sprintf("replay:ban:%s", ip)).Val() == 1 {
			utils.RecordReplayFail(c, "banned_ip", uid, false)
			c.JSON(http.StatusOK, utils.Error(1, "请求过于频繁，请稍后再试"))
			c.Abort()
			return
		}

		// 1. 必需头存在性（仅日志，不触发封禁，避免误伤配置不当的客户端）
		if timestampStr == "" || nonce == "" {
			utils.RecordReplayFail(c, "missing_headers", uid, false)
			c.JSON(http.StatusOK, utils.Error(1, "缺少防重放攻击请求头"))
			c.Abort()
			return
		}

		// 2. nonce 格式校验，避免超长 key 与注入
		if len(nonce) < 8 || len(nonce) > 128 {
			utils.RecordReplayFail(c, "invalid_nonce", uid, true)
			c.JSON(http.StatusOK, utils.Error(1, "无效的请求随机数"))
			c.Abort()
			return
		}

		// 3. 时间戳新鲜度
		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			utils.RecordReplayFail(c, "invalid_timestamp", uid, true)
			c.JSON(http.StatusOK, utils.Error(1, "无效的时间戳"))
			c.Abort()
			return
		}
		now := time.Now().UnixMilli()
		windowMs := replayWindow.Milliseconds()
		if now-timestamp > windowMs || timestamp-now > windowMs {
			utils.RecordReplayFail(c, "expired_timestamp", uid, true)
			c.JSON(http.StatusOK, utils.Error(1, "请求已过期，请重新发送"))
			c.Abort()
			return
		}

		// 4. 已认证接口必须校验请求签名
		if hasAuth {
			if signature == "" {
				utils.RecordReplayFail(c, "missing_signature", uid, true)
				c.JSON(http.StatusOK, utils.Error(1, "缺少请求签名"))
				c.Abort()
				return
			}

			signKey, err := utils.Redis.Get(utils.Ctx, fmt.Sprintf("signkey:%d", uid)).Result()
			if err != nil {
				utils.RecordReplayFail(c, "invalid_signkey", uid, true)
				c.JSON(http.StatusOK, utils.Error(1, "签名密钥无效或已过期"))
				c.Abort()
				return
			}

			bodyHash := ""
			// multipart/form-data 请求体包含随机 boundary，前后端难以一致哈希，按空串参与签名。
			// 上传内容不受该签名保护，建议在业务层对文件内容做哈希/大小/类型校验，并对上传做幂等。
			if strings.HasPrefix(c.ContentType(), "multipart/form-data") {
				bodyHash = utils.Sha256("")
			} else {
				bodyBytes, err := c.GetRawData()
				if err != nil {
					utils.RecordReplayFail(c, "read_body_failed", uid, false)
					c.JSON(http.StatusOK, utils.Error(1, "读取请求体失败"))
					c.Abort()
					return
				}
				// 恢复请求体，供后续中间件和 handler 读取
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				bodyHash = utils.Sha256(string(bodyBytes))
			}

			if !utils.VerifyRequest(signKey, c.Request.Method, c.Request.URL.Path, timestampStr, nonce, bodyHash, signature) {
				utils.RecordReplayFail(c, "invalid_signature", uid, true)
				c.JSON(http.StatusOK, utils.Error(1, "请求签名无效"))
				c.Abort()
				return
			}
		}

		// 5. nonce 一次性使用，绑定用户，避免跨用户/跨会话碰撞
		nonceKey := fmt.Sprintf("replay:nonce:anon:%s", nonce)
		if hasAuth {
			nonceKey = fmt.Sprintf("replay:nonce:%d:%s", uid, nonce)
		}
		ok, err := utils.Redis1.SetNX(utils.Ctx, nonceKey, "1", replayWindow).Result()
		if err != nil {
			utils.RecordReplayFail(c, "nonce_store_failed", uid, false)
			c.JSON(http.StatusOK, utils.Error(1, "请求验证失败"))
			c.Abort()
			return
		}
		if !ok {
			utils.RecordReplayFail(c, "replayed_nonce", uid, true)
			c.JSON(http.StatusOK, utils.Error(1, "检测到重放攻击，请求已被拒绝"))
			c.Abort()
			return
		}

		c.Next()
	}
}
