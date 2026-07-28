package captcha

import (
	"member/pkg/redis"
	"time"

	"github.com/mojocn/base64Captcha"
)

const captchaPrefix = "captcha:"
const captchaExpire = 5 * time.Minute

func Generate() (string, string, error) {
	driver := base64Captcha.NewDriverDigit(80, 240, 4, 0.7, 80)
	c := base64Captcha.NewCaptcha(driver, &redisStore{})
	id, b64s, _, err := c.Generate()
	return id, b64s, err
}

func Verify(id, answer string) bool {
	if id == "" || answer == "" {
		return false
	}
	key := captchaPrefix + id
	stored, err := redis.CaptchaClient.Get(redis.Ctx, key).Result()
	if err != nil {
		return false
	}
	redis.CaptchaClient.Del(redis.Ctx, key)
	return stored == answer
}

type redisStore struct{}

func (s *redisStore) Set(id string, value string) error {
	return redis.CaptchaClient.Set(redis.Ctx, captchaPrefix+id, value, captchaExpire).Err()
}

func (s *redisStore) Get(id string, clear bool) string {
	key := captchaPrefix + id
	val, err := redis.CaptchaClient.Get(redis.Ctx, key).Result()
	if err != nil {
		return ""
	}
	if clear {
		redis.CaptchaClient.Del(redis.Ctx, key)
	}
	return val
}

func (s *redisStore) Verify(id, answer string, clear bool) bool {
	val := s.Get(id, clear)
	return val == answer
}
