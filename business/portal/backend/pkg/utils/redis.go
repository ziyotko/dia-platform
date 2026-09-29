package utils

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"

	"portal/config"
)

var Redis *redis.Client
var Redis1 *redis.Client
var Redis2 *redis.Client

// RedisMember 外部会员（business/member）令牌黑名单库：**只读**，不能写入。
// 用途：member 登出后把 jti 写进它自己的 captcha 库（键 `blacklist:<jti>`），
// portal 读同一个库即可让「会员登出后 portal 侧立即失效」。
// 未启用（redis.member_token_db < 0）时为 nil，鉴权中间件会跳过该检查。
var RedisMember *redis.Client

var Ctx = context.Background()

func InitRedisCaptcha() {
	conf := config.AppConfig.Redis
	Redis = redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.CaptchaDB,
	})

	_, err := Redis.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect Redis: %s", err)
	}
}

// InitRedisMember 连接外部会员（business/member）令牌黑名单所在库。
// 该库属于 member 项目（member 的 redis.captcha_db），portal 只读不写。
// member_token_db < 0 时按「未启用」处理（RedisMember 保持 nil，中间件跳过黑名单检查）。
// 与其它 Redis 初始化一致：连不上直接退出（fail-closed，不能退化成「不检查黑名单」）。
func InitRedisMember() {
	conf := config.AppConfig.Redis
	if conf.MemberTokenDB < 0 {
		log.Printf("[WARN] 未启用外部会员令牌黑名单检查（redis.member_token_db < 0）")
		return
	}
	addr := conf.MemberTokenAddr
	if addr == "" {
		addr = conf.Addr
	}
	password := conf.MemberTokenPassword
	if password == "" {
		password = conf.Password
	}
	RedisMember = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       conf.MemberTokenDB,
	})

	if _, err := RedisMember.Ping(Ctx).Result(); err != nil {
		log.Fatalf("Failed to connect Redis (member token db %d): %s", conf.MemberTokenDB, err)
	}
}
func InitRedisAnti() {
	conf := config.AppConfig.Redis
	Redis1 = redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.AntiReplayDB,
	})
	_, err := Redis1.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect Redis: %s", err)
	}
}

func InitRedisCache() {
	conf := config.AppConfig.Redis
	Redis2 = redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.CacheDB,
	})
	_, err := Redis2.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect Redis: %s", err)
	}
}
