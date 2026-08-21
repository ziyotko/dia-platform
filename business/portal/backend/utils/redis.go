package utils

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"

	"server/config"
)

var Redis *redis.Client
var Redis1 *redis.Client
var Redis2 *redis.Client
var Ctx = context.Background()

func InitRedisCaptcha() {
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

func InitRedisAnti() {
	conf := config.AppConfig.Redis
	Redis1 = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", conf.Host, conf.Port),
		Password: conf.Password,
		DB:       conf.DB1,
	})
	_, err := Redis1.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect Redis: %s", err)
	}
}

func InitRedisCache() {
	conf := config.AppConfig.Redis
	Redis2 = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", conf.Host, conf.Port),
		Password: conf.Password,
		DB:       conf.DB2,
	})
	_, err := Redis2.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect Redis: %s", err)
	}
}
