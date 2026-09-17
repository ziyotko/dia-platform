package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"base/pkg/redis"

	"github.com/mojocn/base64Captcha"
)

const (
	CaptchaKeyPrefix   = "captcha:"
	CaptchaExpire      = 5 * time.Minute
	LoginFailKeyPrefix = "login_fail:"
)

// 数字 + 字母验证码字符集：5 位，仅使用大写字母（与 portal / member / application 保持一致）。
// 注意：
//   - 元素个数必须大于 Length，否则库会回退为默认字符集；
//   - 已剔除易混字符 0/O、1/I/L、2/Z、5/S、8/B，降低识别成本；
//   - 校验时大小写不敏感，用户输入小写字母同样可通过。
const captchaSource = "234679ACDEFGHJKMNPQRTUVWXY"

type CaptchaService struct{}

// redisCaptchaStore 以 Redis 作为验证码存储，多实例部署下校验结果一致。
type redisCaptchaStore struct{}

func (redisCaptchaStore) Set(id string, value string) error {
	return redis.Client.Set(redis.Ctx, CaptchaKeyPrefix+id, value, CaptchaExpire).Err()
}

func (redisCaptchaStore) Get(id string, clear bool) string {
	key := CaptchaKeyPrefix + id
	val, err := redis.Client.Get(redis.Ctx, key).Result()
	if err != nil {
		return ""
	}
	if clear {
		_ = redis.Client.Del(redis.Ctx, key).Err()
	}
	return val
}

func (s redisCaptchaStore) Verify(id, answer string, clear bool) bool {
	val := s.Get(id, clear)
	if val == "" {
		return false
	}
	// 大小写不敏感，兼容用户输入小写字母；同时容忍首尾空格。
	return strings.EqualFold(val, strings.TrimSpace(answer))
}

// Generate 生成验证码：5 位数字 + 大写字母，仅保留少量噪点与细干扰线（与 portal 保持一致的样式与字符集）。
func (s CaptchaService) Generate() (id string, b64s string, err error) {
	// 尺寸保持 100x300（宽:高 = 3:1），契合前端 150x50 的 object-fit 容器，避免裁切。
	driver := base64Captcha.NewDriverString(
		100, 300, // height, width
		// noiseCount 噪点（拉丁字符笔画细，噪点过多会明显影响可读性）
		8,
		// 干扰线只保留 1px 细斜线（OptionShowSlimeLine），降低对文字的遮挡：
		//   - OptionShowHollowLine：绘制 height/20 粗的实心正弦带，遮挡笔画最严重；
		//   - OptionShowSineLine：每列填充 height/5 个像素，形成很宽的波浪带。
		// 两者均已移除，仅保留最轻的细线，兼顾可读性与防 OCR。
		base64Captcha.OptionShowSlimeLine,
		5,             // Length 字符个数
		captchaSource, // 字符集（数字 + 大写字母，剔除易混字符）
		nil,           // 背景颜色（随机浅色）
		nil,           // 字体存储（默认 DefaultEmbeddedFonts）
		nil,
	)
	c := base64Captcha.NewCaptcha(driver, redisCaptchaStore{})
	id, b64s, _, err = c.Generate()
	if err != nil {
		return "", "", err
	}
	return id, b64s, nil
}

// Verify 校验验证码：一次性使用（校验后立即删除），大小写不敏感。
func (s CaptchaService) Verify(id, code string) bool {
	// 空参数直接失败，避免 strings.EqualFold("", "") 误判为通过
	if id == "" || code == "" {
		return false
	}
	return redisCaptchaStore{}.Verify(id, code, true)
}

func loginFailKey(tenantID uint64, username string) string {
	return LoginFailKeyPrefix + strconv.FormatUint(tenantID, 10) + ":" + username
}

// RecordLoginFail 记录一次登录失败。
// 计数键 TTL = 锁定时长：即 maxFail 次失败需在 lockMinutes 窗口内累计，达到阈值后锁定期内持续拦截。
func (s CaptchaService) RecordLoginFail(tenantID uint64, username string, maxFail, lockMinutes int) (int, error) {
	key := loginFailKey(tenantID, username)
	count, err := redis.Client.Incr(redis.Ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 || int(count) == maxFail {
		_ = redis.Client.Expire(redis.Ctx, key, time.Duration(lockMinutes)*time.Minute).Err()
	}
	return int(count), nil
}

// CheckAndLock 校验账号是否因连续登录失败被锁定，锁定时返回剩余等待分钟数。
func (s CaptchaService) CheckAndLock(tenantID uint64, username string, maxFail, lockMinutes int) error {
	key := loginFailKey(tenantID, username)
	count, err := redis.Client.Get(redis.Ctx, key).Int()
	if err != nil || count < maxFail {
		return nil
	}

	remain := lockMinutes
	if ttl, ttlErr := redis.Client.TTL(redis.Ctx, key).Result(); ttlErr == nil && ttl > 0 {
		remain = int(ttl.Minutes()) + 1
	}
	return fmt.Errorf("登录失败次数过多，账号已锁定，请 %d 分钟后重试", remain)
}

func (s CaptchaService) ClearLoginFail(tenantID uint64, username string) error {
	return redis.Client.Del(redis.Ctx, loginFailKey(tenantID, username)).Err()
}
