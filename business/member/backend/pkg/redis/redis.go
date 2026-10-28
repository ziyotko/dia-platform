package redis

import (
	"context"
	"log"
	"member/config"
	"sync"

	goredis "github.com/go-redis/redis/v8"
)

var (
	CaptchaClient    *goredis.Client
	AntiReplayClient *goredis.Client
	Ctx              = context.Background()
)

// newClient 按配置构造客户端。超时/连接池为 0 时交给 go-redis 用自己的默认值，
// 避免把「未配置」误当成「立即超时」。
func newClient(cfg *config.RedisConfig, db int) *goredis.Client {
	return goredis.NewClient(&goredis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           db,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolSize:     cfg.PoolSize,
	})
}

func Init(cfg *config.RedisConfig) {
	CaptchaClient = newClient(cfg, cfg.CaptchaDB)
	AntiReplayClient = newClient(cfg, cfg.AntiReplayDB)

	// 两个 DB 的连通性检查互不依赖：并行 Ping，省掉一次串行网络往返（启动更快）
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if _, err := CaptchaClient.Ping(Ctx).Result(); err != nil {
			log.Fatalf("Failed to connect to captcha Redis: %v", err)
		}
	}()
	go func() {
		defer wg.Done()
		if _, err := AntiReplayClient.Ping(Ctx).Result(); err != nil {
			log.Fatalf("Failed to connect to anti-replay Redis: %v", err)
		}
	}()
	wg.Wait()

	log.Println("Redis connections established")
}
