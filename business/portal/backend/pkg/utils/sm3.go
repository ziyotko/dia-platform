package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/emmansun/gmsm/sm3"
)

func SM3Hash(data string) string {
	sum := sm3.Sum([]byte(data))
	return hex.EncodeToString(sum[:])
}

func SM3HashWithSalt(data string, salt string) string {
	return SM3Hash(data + salt)
}

func GenerateSalt() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// passwordIterations 口令哈希迭代次数（PBKDF2 式的口令拉伸）。
//
// 单轮 SM3 的离线爆破成本极低（GPU 可每秒尝试数十亿次），因此对「口令」必须加工作因子：
// 本实现在盐值参与下重复迭代 passwordIterations 次，把每次校验的代价抬到毫秒级
// （单次登录多花几十毫秒，对暴力破解则是几个数量级的成本提升）。
// 该值写入哈希串（salt:iter:hash），改大后旧哈希仍可校验，用户下次改密/登录时自动升级。
const passwordIterations = 100000

// passwordHashPrefix 新版哈希串的字段数（salt:iterations:hash）。
const passwordHashFields = 3

func SM3HashPassword(password string) string {
	salt := GenerateSalt()
	hashed := derivePasswordHash(password, salt, passwordIterations)
	return fmt.Sprintf("%s:%d:%s", salt, passwordIterations, hashed)
}

// derivePasswordHash 口令拉伸：h = SM3(password + salt)，随后 h = SM3(h + salt) 重复 iter-1 次。
func derivePasswordHash(password, salt string, iter int) string {
	if iter < 1 {
		iter = 1
	}
	h := SM3HashWithSalt(password, salt)
	for i := 1; i < iter; i++ {
		h = SM3HashWithSalt(h, salt)
	}
	return h
}

// VerifySM3Password 校验口令。兼容三种历史格式：
//   - `salt:iterations:hash`（当前格式，带迭代次数）
//   - `salt:hash`（旧格式：单轮加盐 SM3）
//   - 无分隔符的历史脏值 → 一律视为校验失败（不能用明文比较）
//
// 比较使用常量时间函数，避免通过响应时间区分「哈希前缀正确」与「完全不同」。
func VerifySM3Password(password string, hashedPassword string) bool {
	idx := strings.Index(hashedPassword, ":")
	if idx <= 0 {
		return false
	}
	salt := hashedPassword[:idx]
	rest := hashedPassword[idx+1:]

	iterations := 1
	storedHash := rest
	if sep := strings.Index(rest, ":"); sep > 0 {
		// 新版格式：中间段是迭代次数；解析失败（历史脏值恰好含冒号）时按单轮处理
		if n, err := strconv.Atoi(rest[:sep]); err == nil && n > 0 && n <= passwordIterations*10 {
			iterations = n
			storedHash = rest[sep+1:]
		} else {
			return false
		}
	}
	if storedHash == "" {
		return false
	}

	computedHash := derivePasswordHash(password, salt, iterations)
	return subtle.ConstantTimeCompare([]byte(computedHash), []byte(storedHash)) == 1
}

// PasswordHashNeedsUpgrade 判断已存储的口令哈希是否为旧格式（单轮，无迭代次数），
// 供登录成功后透明升级为带工作因子的新格式使用。
func PasswordHashNeedsUpgrade(hashedPassword string) bool {
	idx := strings.Index(hashedPassword, ":")
	if idx <= 0 {
		return true
	}
	rest := hashedPassword[idx+1:]
	if sep := strings.Index(rest, ":"); sep > 0 {
		n, err := strconv.Atoi(rest[:sep])
		return err != nil || n < passwordIterations
	}
	return true
}
