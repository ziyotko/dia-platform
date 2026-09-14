package utils

import (
	"strings"
	"time"

	"github.com/mojocn/base64Captcha"
)

const captchaPrefix = "captcha:"
const captchaExpire = 5 * time.Minute

// 数字 + 字母验证码字符集：5 位，仅使用大写字母。
// 注意：
//   - 元素个数必须大于 Length，否则库会回退为默认字符集；
//   - 已剔除易混字符 0/O、1/I/L、2/Z、5/S、8/B，降低识别成本；
//   - 校验时大小写不敏感，用户输入小写字母同样可通过。
const captchaSource = "234679ACDEFGHJKMNPQRTUVWXY"

func GenerateCaptcha() (string, string, error) {
	// 数字 + 字母验证码：5 位字符（不含汉字）。
	// 尺寸保持 100x300（宽:高 = 3:1），契合前端 150x50 的 object-fit 容器，避免裁切。
	driver := base64Captcha.NewDriverString(
		100, 300, // height, width
		30, // noiseCount 噪点（拉丁字符笔画细，噪点过多会明显影响可读性）
		base64Captcha.OptionShowHollowLine|base64Captcha.OptionShowSlimeLine|base64Captcha.OptionShowSineLine,
		5,             // Length 字符个数
		captchaSource, // 字符集（数字 + 大写字母，剔除易混字符）
		nil,           // 背景颜色（随机浅色）
		nil,           // 字体存储（默认 DefaultEmbeddedFonts）
		// 传 nil 表示使用内置全部字体，drawText 对每个字符随机取字体，
		// 拉丁字符在所有内置字体中都有字形，不会出现方块。
		nil,
	)
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

	// 大小写不敏感，兼容用户输入小写字母；同时容忍首尾空格。
	return strings.EqualFold(answer, strings.TrimSpace(code))
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
	return strings.EqualFold(val, strings.TrimSpace(answer))
}
