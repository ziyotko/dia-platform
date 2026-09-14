package utils

import (
	"time"

	"github.com/mojocn/base64Captcha"
)

const captchaPrefix = "captcha:"
const captchaExpire = 5 * time.Minute

// 中文验证码字符集：以逗号分隔。
// 注意：元素个数必须大于 Length，否则库会回退为数字/字母；避免使用易混字符。
// 已混入部分不常用字以提升抗爆破/抗OCR能力；若个别字不在字体中会显示为方框，可从此处移除。
const captchaChineseSource = "春,夏,秋,冬,晨,暮,朝,夜,山,水,云,风,雨,雪,雾,霜,露,虹,霞,雷," +
	"天,地,日,月,星,光,影,宇,辰,旭,晓,晖,花,草,木,林,松,竹,梅,兰,菊,荷,桂,柏," +
	"鸟,鱼,鹤,燕,鸿,湖,海,河,江,川,溪,泉,泽,峰,岭,岩,石,畔,涧," +
	"岚,嵩,昊,晟,曦,澈,涵,淇,泓,澜,瑾,璞,珺,璟,鹭,鸾"

func GenerateCaptcha() (string, string, error) {
	// 中文验证码：5 个汉字，混入不常用字，增强抗 OCR 能力（渲染源含 CJK 字体 wqy-microhei）。
	// 尺寸保持 100x300（宽:高 = 3:1），契合前端 120x40 的 object-fit 容器，避免裁切。
	driver := base64Captcha.NewDriverChinese(
		100, 300, // height, width
		90, // noiseCount 噪点
		base64Captcha.OptionShowHollowLine|base64Captcha.OptionShowSlimeLine|base64Captcha.OptionShowSineLine,
		5,                    // Length 汉字个数
		captchaChineseSource, // 字符集
		nil,                  // 背景颜色（随机浅色）
		nil,                  // 字体存储（默认 DefaultEmbeddedFonts）
		// 关键：显式指定只用中文字体 wqy-microhei.ttc。
		// 若传 nil，库会用 fontsAll（含大量纯拉丁字体），而 drawText 对每个字符
		// 都会 randFontFrom 随机选字体，选中拉丁字体时无汉字字形 → 渲染成方块。
		[]string{"wqy-microhei.ttc"},
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

	return answer == code
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
