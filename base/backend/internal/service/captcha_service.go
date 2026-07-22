package service

import (
	"errors"
	"time"

	"base/pkg/redis"

	"github.com/mojocn/base64Captcha"
)

const (
	CaptchaKeyPrefix   = "captcha:"
	CaptchaExpire      = 5 * time.Minute
	LoginFailKeyPrefix = "login_fail:"
	LoginFailExpire    = 15 * time.Minute
	MaxLoginFailCount  = 5
	LockDuration       = 30 * time.Minute
)

type CaptchaService struct{}

var store = base64Captcha.DefaultMemStore

func (s CaptchaService) Generate() (id string, b64s string, err error) {
	driver := base64Captcha.NewDriverDigit(80, 240, 4, 0.7, 80)
	c := base64Captcha.NewCaptcha(driver, store)
	id, b64s, answer, err := c.Generate()
	if err != nil {
		return "", "", err
	}
	// 同时写入 redis，便于集群部署时统一校验
	_ = redis.Client.Set(redis.Ctx, CaptchaKeyPrefix+id, answer, CaptchaExpire).Err()
	return id, b64s, nil
}

func (s CaptchaService) Verify(id, code string) bool {
	if id == "" || code == "" {
		return false
	}
	// 优先使用 redis 校验
	val, err := redis.Client.Get(redis.Ctx, CaptchaKeyPrefix+id).Result()
	if err == nil && val != "" {
		ok := val == code
		if ok {
			_ = redis.Client.Del(redis.Ctx, CaptchaKeyPrefix+id).Err()
		}
		return ok
	}
	// 兜底使用内存 store
	return store.Verify(id, code, true)
}

func (s CaptchaService) RecordLoginFail(username string) (int, error) {
	key := LoginFailKeyPrefix + username
	count, err := redis.Client.Incr(redis.Ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		_ = redis.Client.Expire(redis.Ctx, key, LoginFailExpire).Err()
	}
	return int(count), nil
}

func (s CaptchaService) IsLocked(username string) (bool, int, error) {
	key := LoginFailKeyPrefix + username
	count, err := redis.Client.Get(redis.Ctx, key).Int()
	if err != nil {
		// redis 中没有失败记录，未锁定，剩余次数为最大允许次数
		return false, MaxLoginFailCount, nil
	}
	if count >= MaxLoginFailCount {
		return true, 0, nil
	}
	return false, MaxLoginFailCount - count, nil
}

func (s CaptchaService) ClearLoginFail(username string) error {
	return redis.Client.Del(redis.Ctx, LoginFailKeyPrefix+username).Err()
}

func (s CaptchaService) CheckAndLock(username string) error {
	locked, remain, err := s.IsLocked(username)
	if err != nil {
		return err
	}
	if locked || remain <= 0 {
		return errors.New("登录失败次数过多，账号已锁定，请30分钟后再试")
	}
	return nil
}
