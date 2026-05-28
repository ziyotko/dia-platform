package utils

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/mojocn/base64Captcha"
)

const captchaPrefix = "captcha:"
const captchaExpire = 5 * time.Minute

func GenerateCaptcha() (string, string, error) {
	driver := base64Captcha.NewDriverDigit(80, 240, 4, 0.7, 80)
	captcha := base64Captcha.NewCaptcha(driver, redisCaptchaStore{})

	captchaId, b64s, answer, err := captcha.Generate()
	if err != nil {
		return "", "", err
	}

	err = Redis.Set(Ctx, captchaPrefix+captchaId, answer, captchaExpire).Err()
	if err != nil {
		return "", "", err
	}

	return captchaId, b64s, nil
}

func VerifyCaptcha(id, code string) bool {
	if id == "" || code == "" {
		return false
	}

	key := captchaPrefix + id
	answer, err := Redis.Get(Ctx, key).Result()
	if err != nil {
		return false
	}

	err = Redis.Del(Ctx, key).Err()
	if err != nil {
		return false
	}

	return answer == code
}

func GenerateSmsCaptcha() (string, error) {
	b := make([]byte, 3)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	num := (int(b[0])*256*256 + int(b[1])*256 + int(b[2])) % 1000000
	return fmt.Sprintf("%06d", num), nil
}

func StoreSmsCaptcha(mobile, code string) error {
	key := captchaPrefix + "sms:" + mobile
	return Redis.Set(Ctx, key, code, captchaExpire).Err()
}

func VerifySmsCaptcha(mobile, code string) bool {
	if mobile == "" || code == "" {
		return false
	}

	key := captchaPrefix + "sms:" + mobile
	storedCode, err := Redis.Get(Ctx, key).Result()
	if err != nil {
		return false
	}

	err = Redis.Del(Ctx, key).Err()
	if err != nil {
		return false
	}

	return storedCode == code
}

type redisCaptchaStore struct{}

func (r redisCaptchaStore) Set(id string, value string) error {
	return Redis.Set(Ctx, captchaPrefix+id, value, captchaExpire).Err()
}

func (r redisCaptchaStore) Get(id string, clear bool) string {
	key := captchaPrefix + id
	val, err := Redis.Get(Ctx, key).Result()
	if err != nil {
		return ""
	}
	if clear {
		Redis.Del(Ctx, key)
	}
	return val
}

func (r redisCaptchaStore) Verify(id, answer string, clear bool) bool {
	val := r.Get(id, clear)
	return val == answer
}
