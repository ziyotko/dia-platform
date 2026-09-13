package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"server/config"
	"server/utils"
)

// 仅对不超过该大小的 multipart 请求体在服务端重算文件哈希，避免超大视频上传导致内存压力。
const maxMultipartHashVerify = 20 << 20

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

			// 签名密钥由 Token 派生（不再依赖 Redis 存储的 signKey）；
			// Token 已在 AuthMiddleware 完成校验，这里只需从 Authorization 头提取。
			signKey := utils.DeriveSignKey(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))

			bodyHash := ""
			isMultipart := strings.HasPrefix(c.ContentType(), "multipart/form-data")
			if isMultipart {
				// 上传请求体含随机 boundary，无法直接对原始字节一致哈希。
				// 前端计算文件内容哈希并通过 X-Body-Hash-Value 上报，参与签名；
				// 若未上报（旧客户端/超大文件），回退为空串哈希以保持兼容。
				if h := c.GetHeader("X-Body-Hash-Value"); h != "" {
					bodyHash = h
				} else {
					bodyHash = utils.Sha256("")
				}
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

			// 签名覆盖 path + query，防止 GET 查询参数被篡改
			if !utils.VerifyRequest(signKey, c.Request.Method, c.Request.URL.RequestURI(), timestampStr, nonce, bodyHash, signature) {
				utils.RecordReplayFail(c, "invalid_signature", uid, true)
				c.JSON(http.StatusOK, utils.Error(1, "请求签名无效"))
				c.Abort()
				return
			}

			// multipart 上传：服务端重算文件内容哈希，与声明值比对，提供真实的内容完整性校验
			if isMultipart {
				if !verifyMultipartBodyHash(c, c.GetHeader("X-Body-Hash-Value")) {
					utils.RecordReplayFail(c, "multipart_hash_mismatch", uid, true)
					c.JSON(http.StatusOK, utils.Error(1, "上传文件内容校验失败"))
					c.Abort()
					return
				}
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

// verifyMultipartBodyHash 重算 multipart 请求体中各文件内容（按上传顺序拼接）的 SHA-256，
// 与客户端上报的 X-Body-Hash-Value 比对，从而对上传文件内容提供真实的完整性校验。
//
// 规则：
//   - Content-Length 未知或超过 maxMultipartHashVerify 时跳过（超大文件无法在内存中重算），
//     此时仍依赖签名绑定声明哈希 + 控制器层的文件类型/大小校验；
//   - 仅含普通字段、不含文件的 multipart 直接放行（无内容可校验）；
//   - 含文件但未上报 X-Body-Hash-Value 时**拒绝**，防止客户端“降级”跳过文件完整性校验。
func verifyMultipartBodyHash(c *gin.Context, declared string) bool {
	if cl := c.Request.ContentLength; cl == -1 || cl > maxMultipartHashVerify {
		return true
	}

	bodyBytes, err := c.GetRawData()
	if err != nil {
		return false
	}
	// 恢复请求体，供后续中间件和 controller 读取（如 FormFile / ParseMultipartForm）
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	_, params, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
	if err != nil || params["boundary"] == "" {
		return false
	}

	h := sha256.New()
	hasFile := false
	reader := multipart.NewReader(bytes.NewReader(bodyBytes), params["boundary"])
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false
		}
		// 仅对文件部分（含 filename）计算内容哈希，与前端 FormData 顺序保持一致
		if part.FileName() != "" {
			hasFile = true
			if _, err := io.Copy(h, part); err != nil {
				return false
			}
		}
		part.Close()
	}

	// 无文件的 multipart：无内容可校验，放行
	if !hasFile {
		return true
	}
	// 含文件但缺失声明哈希：拒绝，避免通过不带上报头绕过内容校验
	if declared == "" {
		return false
	}
	return hex.EncodeToString(h.Sum(nil)) == declared
}
