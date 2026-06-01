package utils

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"

	"server/config"
)

var Redis *redis.Client
var Ctx = context.Background()

func InitRedis() {
	conf := config.AppConfig.Redis
	Redis = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", conf.Host, conf.Port),
		Password: conf.Password,
		DB:       conf.DB,
	})

	_, err := Redis.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect Redis: %s", err)
	}
}
