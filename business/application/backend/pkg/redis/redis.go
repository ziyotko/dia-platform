package redis

import (
	"context"
	"time"

	"application/config"

	goredis "github.com/go-redis/redis/v8"
)

var (
	Ctx              = context.Background()
	CaptchaClient    *goredis.Client
	AntiReplayClient *goredis.Client
)

// BlacklistKeyPrefix 登出黑名单键前缀。值固定写 "1"（只判存在性，不存 Token 原文：
// 否则「能读 Redis」就等同于「拿到可用凭证」）。存放在 captcha 库（与验证码同库，
// 都是短生命周期会话态数据）。
const BlacklistKeyPrefix = "blacklist:"

// logoutBlacklistTTLFallback Token 缺少过期时间时的兜底 TTL（正常不会用到）
const logoutBlacklistTTLFallback = 24 * time.Hour

// RevokeToken 把 Token 的 jti 写入黑名单，使其在自然过期前不再可用（登出用）。
// expiresAt 为零值或已过期时不写（Token 本已失效）。
func RevokeToken(jti string, expiresAt time.Time) error {
	if jti == "" {
		return nil
	}
	ttl := logoutBlacklistTTLFallback
	if !expiresAt.IsZero() {
		ttl = time.Until(expiresAt)
		if ttl <= 0 {
			return nil
		}
	}
	return CaptchaClient.Set(Ctx, BlacklistKeyPrefix+jti, "1", ttl).Err()
}

// IsTokenRevoked 判断 jti 是否已在登出黑名单中。
// Redis 异常时返回 (true, err)：调用方按 fail-closed 处理，避免缓存故障变成鉴权放开。
func IsTokenRevoked(jti string) (bool, error) {
	if jti == "" {
		return false, nil
	}
	_, err := CaptchaClient.Get(Ctx, BlacklistKeyPrefix+jti).Result()
	if err == nil {
		return true, nil
	}
	if err == goredis.Nil {
		return false, nil
	}
	return true, err
}

func Init(cfg *config.RedisConfig) {
	CaptchaClient = goredis.NewClient(&goredis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.CaptchaDB,
	})
	if _, err := CaptchaClient.Ping(Ctx).Result(); err != nil {
		panic("Failed to connect Redis (captcha): " + err.Error())
	}

	AntiReplayClient = goredis.NewClient(&goredis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.AntiReplayDB,
	})
	if _, err := AntiReplayClient.Ping(Ctx).Result(); err != nil {
		panic("Failed to connect Redis (anti-replay): " + err.Error())
	}
}
