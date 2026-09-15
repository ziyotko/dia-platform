package service

import (
	"errors"
	"strconv"
	"time"

	"base/config"
	"base/pkg/redis"

	goredis "github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// TokenService JWT 吊销（登出 / 改密），存储全部放在 Redis，不新增数据表：
//
//   - auth:token:revoked:<jti>          单次登出：该 token 进黑名单，TTL = 剩余有效期
//   - auth:user:revoked-before:<userID> 用户级失效：该时刻之前签发的 token 全部失效（改密等）
//
// Redis 不可用时 fail-open（放行并记日志）：宁可短暂失去吊销能力，也不能让所有人无法登录。
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

// IsTokenRevoked 判断 token 是否已被登出吊销。Redis 异常时返回 false（fail-open）。
func (s TokenService) IsTokenRevoked(tokenID string) bool {
	if tokenID == "" {
		return false
	}
	n, err := redis.Client.Exists(redis.Ctx, tokenRevokedPrefix+tokenID).Result()
	if err != nil {
		logrus.WithError(err).Warn("校验 token 黑名单失败，已放行")
		return false
	}
	return n > 0
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
func (s TokenService) UserRevokedBefore(userID uint64) (time.Time, bool) {
	raw, err := redis.Client.Get(redis.Ctx, userRevokedBeforeKey(userID)).Result()
	if err != nil {
		if !errors.Is(err, goredis.Nil) {
			logrus.WithError(err).Warn("读取用户 token 失效时间失败，已放行")
		}
		return time.Time{}, false
	}
	sec, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return time.Time{}, false
	}
	return time.Unix(sec, 0), true
}

func userRevokedBeforeKey(userID uint64) string {
	return userRevokedBeforePref + strconv.FormatUint(userID, 10)
}
