package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"base/config"
	"base/pkg/redis"

	goredis "github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// RefreshTokenService 会话续期用的 refresh token（不透明随机串，库里只存 SHA-256 摘要）。
//
// 设计要点（OAuth 2.0 安全最佳实践的精简落地）：
//   - **轮换**：每次续期都作废旧 refresh token 并签发新的，旧串立即失效；
//   - **复用检测**：已被消费过的 refresh token 再次出现 → 判定为泄漏，直接吊销该用户全部会话
//     （改密/禁用/删除用户走同一套 `auth:user:revoked-before` 机制，refresh token 一并失效）；
//   - **宽限期重放**：刚轮换后的 120 秒内用同一个串重放，返回**同一对**新令牌，
//     避免前端并发/网络重试被误判为泄漏（前端另有单飞控制，双保险）；
//   - **fail-closed**：Redis 不可用时拒绝续期（无法校验的凭证不能放行），用户仍可继续用完当前 access token。
//
// 存储（全部在 Redis，不新增数据表）：
//
//	auth:refresh:<sha256>      有效记录 {userId, issuedAt}，TTL = jwt.refresh_expire_hours
//	auth:refresh:rev:<sha256>  已消费/已作废标记（userId），TTL 同上 → 再次出现即复用
//	auth:refresh:used:<sha256> 宽限期重放缓存（本次签发的那一对令牌），TTL 120s
const (
	refreshKeyPrefix   = "auth:refresh:"
	refreshRevPrefix   = "auth:refresh:rev:"
	refreshUsedPrefix  = "auth:refresh:used:"
	refreshGraceWindow = 120 * time.Second
)

var (
	// errRefreshInvalid token 无效/已失效（不存在、已登出、用户级吊销之后签发）
	errRefreshInvalid = errors.New("登录状态已失效，请重新登录")
	// errRefreshStorage Redis 不可用（fail-closed）
	errRefreshStorage = errors.New("服务暂时不可用，请稍后重试")
	// errRefreshReuse 复用检测命中
	errRefreshReuse = errors.New("登录状态异常，请重新登录")
)

// TokenPair 一次签发的令牌对（refresh_token 只在响应体里返回，库里只有摘要）。
type TokenPair struct {
	AccessToken  string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // access token 有效期（秒）
}

type RefreshTokenService struct{}

// RefreshTokenTTL refresh token 有效期（jwt.refresh_expire_hours，默认 168 小时 = 7 天）。
func RefreshTokenTTL() time.Duration {
	hours := 0
	if config.Cfg != nil {
		hours = config.Cfg.JWT.RefreshExpireHours
	}
	if hours <= 0 {
		hours = 168
	}
	return time.Duration(hours) * time.Hour
}

type refreshRecord struct {
	UserID   uint64 `json:"userId"`
	IssuedAt int64  `json:"issuedAt"`
}

// newOpaqueToken 生成 256 位随机串（hex，64 字符）。
func newOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func sha256Hex(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// Issue 为用户签发 refresh token。返回的明文只在本次响应里使用。
func (s RefreshTokenService) Issue(userID uint64) (string, error) {
	token, err := newOpaqueToken()
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(refreshRecord{UserID: userID, IssuedAt: time.Now().Unix()})
	if err != nil {
		return "", err
	}
	if err := redis.Client.Set(redis.Ctx, refreshKeyPrefix+sha256Hex(token), payload, RefreshTokenTTL()).Err(); err != nil {
		return "", err
	}
	return token, nil
}

// Load 校验并解析 refresh token，返回用户 ID。
// 除「记录存在」外，还要求签发时间晚于该用户的 token 失效时间点
// （改密/禁用/删除用户会写 auth:user:revoked-before），因此那些操作会自动让 refresh token 失效。
func (s RefreshTokenService) Load(token string) (uint64, error) {
	if token == "" {
		return 0, errRefreshInvalid
	}
	raw, err := redis.Client.Get(redis.Ctx, refreshKeyPrefix+sha256Hex(token)).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return 0, errRefreshInvalid
		}
		logrus.WithError(err).Warn("读取 refresh token 失败，拒绝续期")
		return 0, errRefreshStorage
	}
	var rec refreshRecord
	if err := json.Unmarshal([]byte(raw), &rec); err != nil {
		return 0, errRefreshInvalid
	}
	if before, ok := (TokenService{}).UserRevokedBefore(rec.UserID); ok && time.Unix(rec.IssuedAt, 0).Before(before) {
		return 0, errRefreshInvalid
	}
	return rec.UserID, nil
}

// Consume 标记旧 token 已消费：删除有效记录、写长期复用标记、缓存本次签发的一对令牌供宽限期重放。
func (s RefreshTokenService) Consume(oldToken string, userID uint64, pair TokenPair) error {
	h := sha256Hex(oldToken)
	pipe := redis.Client.Pipeline()
	pipe.Del(redis.Ctx, refreshKeyPrefix+h)
	pipe.Set(redis.Ctx, refreshRevPrefix+h, userID, RefreshTokenTTL())
	if payload, err := json.Marshal(pair); err == nil {
		pipe.Set(redis.Ctx, refreshUsedPrefix+h, payload, refreshGraceWindow)
	}
	_, err := pipe.Exec(redis.Ctx)
	return err
}

// Replay 宽限期内用同一个 refresh token 重放时，返回上次签发的那一对（幂等，避免误判泄漏）。
func (s RefreshTokenService) Replay(oldToken string) (TokenPair, bool) {
	if oldToken == "" {
		return TokenPair{}, false
	}
	raw, err := redis.Client.Get(redis.Ctx, refreshUsedPrefix+sha256Hex(oldToken)).Result()
	if err != nil {
		return TokenPair{}, false
	}
	var pair TokenPair
	if err := json.Unmarshal([]byte(raw), &pair); err != nil || pair.AccessToken == "" {
		return TokenPair{}, false
	}
	return pair, true
}

// ReuseDetected 判断该 token 是否是「已经被消费过」的（宽限期之外再次出现即视为泄漏）。
func (s RefreshTokenService) ReuseDetected(token string) (uint64, bool) {
	if token == "" {
		return 0, false
	}
	raw, err := redis.Client.Get(redis.Ctx, refreshRevPrefix+sha256Hex(token)).Result()
	if err != nil {
		return 0, false
	}
	userID, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0, false
	}
	return userID, true
}

// Revoke 主动作废（登出）：删除有效记录并留复用标记，避免登出后的 refresh token 还能续期。
func (s RefreshTokenService) Revoke(token string, userID uint64) error {
	if token == "" {
		return nil
	}
	h := sha256Hex(token)
	pipe := redis.Client.Pipeline()
	pipe.Del(redis.Ctx, refreshKeyPrefix+h)
	if userID > 0 {
		pipe.Set(redis.Ctx, refreshRevPrefix+h, userID, RefreshTokenTTL())
	}
	_, err := pipe.Exec(redis.Ctx)
	return err
}
