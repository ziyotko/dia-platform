package utils

import (
	"crypto/rand"
	"math/big"
)

// passwordCharset 随机初始密码字符集：大小写字母 + 数字 + 常见特殊字符。
// 去掉了易混淆的 0/O、1/l/I，便于管理员口头/短信转告。
const passwordCharset = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789!@#$%^&*"

// defaultRandomPasswordLength 随机密码缺省长度。
const defaultRandomPasswordLength = 12

// RandomPassword 用 crypto/rand 生成 n 位随机密码（n<=0 时取 12 位）。
//
// 用途：管理员重置会员密码 / 新增会员未填密码时生成的初始密码。
// 之前这两处都使用公开常量 DefaultPassword，等于给账号装上「公开钥匙」。
func RandomPassword(n int) (string, error) {
	if n <= 0 {
		n = defaultRandomPasswordLength
	}
	buf := make([]byte, n)
	max := big.NewInt(int64(len(passwordCharset)))
	for i := range buf {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		buf[i] = passwordCharset[idx.Int64()]
	}
	return string(buf), nil
}
