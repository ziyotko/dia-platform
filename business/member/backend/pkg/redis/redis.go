package redis

import (
	"context"
	"log"
	"member/config"

	goredis "github.com/go-redis/redis/v8"
)

var (
	CaptchaClient    *goredis.Client
	AntiReplayClient *goredis.Client
	Ctx              = context.Background()
)

func Init(cfg *config.RedisConfig) {
	CaptchaClient = goredis.NewClient(&goredis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.CaptchaDB,
	})
	if _, err := CaptchaClient.Ping(Ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to captcha Redis: %v", err)
	}

	AntiReplayClient = goredis.NewClient(&goredis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.AntiReplayDB,
	})
	if _, err := AntiReplayClient.Ping(Ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to anti-replay Redis: %v", err)
	}

	log.Println("Redis connections established")
}
