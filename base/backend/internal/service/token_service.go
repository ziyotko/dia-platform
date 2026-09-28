package service

import (
	"errors"
	"strconv"
	"time"

	"base/config"
	"base/pkg/redis"

	goredis "github.com/go-redis/redis/v8"
)

// TokenService JWT 吊销（登出 / 改密），存储全部放在 Redis，不新增数据表：
//
//   - auth:token:revoked:<jti>          单次登出：该 token 进黑名单，TTL = 剩余有效期
//   - auth:user:revoked-before:<userID> 用户级失效：该时刻之前签发的 token 全部失效
//     （禁用/删除/改密/refresh token 重用检测均会写；改密另有落库的 password_changed_at 兜底）
//
// 读失败时**不再放行**（fail-closed）：调用方返回业务码 1 + 「稍后重试」，
// 既不会因缓存故障而放行已吊销的 Token，也不会把用户误登出（只有真失效才返回 401）。
// 写失败只告警：账号禁用/删除的正确性由 JWTAuth 的库内存活校验兜底，不依赖 Redis。
type TokenService struct{}

const (
	tokenRevokedPrefix    = "auth:token:revoked:"
	userRevokedBeforePref = "auth:user:revoked-before:"
)

// RevokeToken 把某个 token（jti）加入黑名单，TTL 取其剩余有效期。
func (s TokenService) RevokeToken(tokenID string, expiresAt time.Time) error {
	if tokenID == "" {
		return nil
	}
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		// token 已过期，无需吊销
		return nil
	}
	return redis.Client.Set(redis.Ctx, tokenRevokedPrefix+tokenID, "1", ttl).Err()
}

// IsTokenRevoked 判断 token 是否已被登出吊销。Redis 异常时返回 error，由调用方 fail-closed。
func (s TokenService) IsTokenRevoked(tokenID string) (bool, error) {
	if tokenID == "" {
		return false, nil
	}
	n, err := redis.Client.Exists(redis.Ctx, tokenRevokedPrefix+tokenID).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// RevokeUserTokensBefore 让该用户在 at 之前签发的所有 token 失效（改密等场景）。
// TTL 取 JWT 有效期：等所有旧 token 自然过期后即可自动清理。
func (s TokenService) RevokeUserTokensBefore(userID uint64, at time.Time) error {
	ttl := time.Duration(config.Cfg.JWT.ExpireHours) * time.Hour
	if ttl <= 0 {
		ttl = 8 * time.Hour
	}
	return redis.Client.Set(redis.Ctx, userRevokedBeforeKey(userID), strconv.FormatInt(at.Unix(), 10), ttl).Err()
}

// UserRevokedBefore 读取用户的「token 失效时间点」，不存在表示没有失效记录。
// Redis 异常时返回 error（调用方 fail-closed）；值脏（非数字）当作无记录，不阻断。
func (s TokenService) UserRevokedBefore(userID uint64) (time.Time, bool, error) {
	raw, err := redis.Client.Get(redis.Ctx, userRevokedBeforeKey(userID)).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return time.Time{}, false, nil
		}
		return time.Time{}, false, err
	}
	sec, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return time.Time{}, false, nil
	}
	return time.Unix(sec, 0), true, nil
}

func userRevokedBeforeKey(userID uint64) string {
	return userRevokedBeforePref + strconv.FormatUint(userID, 10)
}
