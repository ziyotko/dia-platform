package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// GenerateSignKey 生成一个 32 字节的随机签名密钥，以 hex 字符串返回
func GenerateSignKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Sha256 计算字符串的 SHA-256 哈希，返回 hex 字符串
func Sha256(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// HmacSha256 计算 HMAC-SHA256
func HmacSha256(data, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// SignRequest 构造请求签名
// payload = METHOD|TARGET|TIMESTAMP|NONCE|SHA256(BODY)
// TARGET 为 URL 的 path + query（与后端 c.Request.URL.RequestURI() 一致），
// 从而让方法、路径、查询参数与请求体一起受到签名保护。
func SignRequest(signKey, method, target, timestamp, nonce, bodyHash string) string {
	payload := fmt.Sprintf("%s|%s|%s|%s|%s", method, target, timestamp, nonce, bodyHash)
	return HmacSha256(payload, signKey)
}

// VerifyRequest 安全地比较请求签名
func VerifyRequest(signKey, method, target, timestamp, nonce, bodyHash, signature string) bool {
	expected := SignRequest(signKey, method, target, timestamp, nonce, bodyHash)
	return hmac.Equal([]byte(expected), []byte(signature))
}
