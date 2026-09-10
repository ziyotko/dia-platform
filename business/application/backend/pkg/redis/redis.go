package redis

import (
	"context"

	"application/config"

	goredis "github.com/go-redis/redis/v8"
)

var (
	Ctx              = context.Background()
	CaptchaClient    *goredis.Client
	AntiReplayClient *goredis.Client
)

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
