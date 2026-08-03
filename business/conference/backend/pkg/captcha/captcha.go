package captcha

import (
	"conference/pkg/redis"

	"github.com/mojocn/base64Captcha"
)

var Store base64Captcha.Store

type redisStore struct{}

func (s *redisStore) Set(id string, value string) error {
	return redis.CaptchaClient.Set(redis.Ctx, "captcha:"+id, value, base64Captcha.Expiration).Err()
}

func (s *redisStore) Get(id string, clear bool) string {
	val, err := redis.CaptchaClient.Get(redis.Ctx, "captcha:"+id).Result()
	if err != nil {
		return ""
	}
	if clear {
		redis.CaptchaClient.Del(redis.Ctx, "captcha:"+id)
	}
	return val
}

func (s *redisStore) Verify(id, answer string, clear bool) bool {
	val := s.Get(id, clear)
	return val != "" && val == answer
}

func Init() {
	Store = &redisStore{}
}

func Generate() (string, string, string, error) {
	driver := base64Captcha.NewDriverDigit(80, 240, 4, 0.7, 80)
	c := base64Captcha.NewCaptcha(driver, Store)
	id, b64s, answer, err := c.Generate()
	return id, b64s, answer, err
}

func Verify(id, answer string) bool {
	return Store.Verify(id, answer, true)
}
