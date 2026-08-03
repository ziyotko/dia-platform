package utils

import (
	"crypto/rand"
	"math/big"
)

// passwordChars 不含易混淆字符（0/O、1/l/I 等）
const passwordChars = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"

// GenerateRandomPassword 生成包含大小写字母与数字的随机密码。
func GenerateRandomPassword(length int) string {
	if length < 8 {
		length = 8
	}
	max := big.NewInt(int64(len(passwordChars)))
	buf := make([]byte, length)
	for i := range buf {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			buf[i] = 'x'
			continue
		}
		buf[i] = passwordChars[n.Int64()]
	}
	return string(buf)
}
