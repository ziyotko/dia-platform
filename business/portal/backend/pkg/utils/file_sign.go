package utils

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	"portal/config"
)

// FileAccessSignTTL 私有文件「短时效签名 URL」的有效期。
// 时间由后端签发时决定（当前业务固定 5 分钟）：URL 即使外泄，过期后即不可用。
const FileAccessSignTTL = 5 * time.Minute

// fileAccessSignKey 私有文件签名密钥：由 JWT 密钥派生（随部署密钥轮换而失效），
// 使用独立派生串，不与访问 Token / 防重放签名共用同一密钥用途。
func fileAccessSignKey() string {
	return HmacSha256(config.AppConfig.JWT.Secret, "portal-file-access-sign-v1")
}

// SignFileAccess 为「资源标识」（当前为文件名）签发短时效签名参数。
//
// payload = 资源标识|过期时间|nonce：三者任一被篡改都会导致校验失败；
// nonce 取随机数，保证**每次下发的 URL 都不同**（同时避免浏览器/代理缓存串用）。
func SignFileAccess(resource string, ttl time.Duration) (exp int64, nonce string, sign string) {
	if ttl <= 0 {
		ttl = FileAccessSignTTL
	}
	exp = time.Now().Add(ttl).Unix()
	nonce = GenerateSalt()
	sign = HmacSha256(fileAccessSignPayload(resource, exp, nonce), fileAccessSignKey())
	return exp, nonce, sign
}

// VerifyFileAccessSign 校验签名：参数缺失 / 已过期 / 签名不符均返回错误。
// 用 constant-time 比较，避免时序侧信道。
func VerifyFileAccessSign(resource string, exp int64, nonce, sign string) error {
	if exp <= 0 || nonce == "" || sign == "" {
		return errors.New("缺少访问签名")
	}
	if time.Now().Unix() > exp {
		return errors.New("访问链接已过期，请刷新后重试")
	}
	expected := HmacSha256(fileAccessSignPayload(resource, exp, nonce), fileAccessSignKey())
	if subtle.ConstantTimeCompare([]byte(expected), []byte(sign)) != 1 {
		return errors.New("访问签名校验失败")
	}
	return nil
}

// fileAccessSignPayload 拼接签名原文
func fileAccessSignPayload(resource string, exp int64, nonce string) string {
	return fmt.Sprintf("%s|%d|%s", resource, exp, nonce)
}
