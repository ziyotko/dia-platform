package utils

import (
	"crypto/rand"
	"encoding/hex"

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

func SM3HashPassword(password string) string {
	salt := GenerateSalt()
	hashed := SM3HashWithSalt(password, salt)
	return salt + ":" + hashed
}

func VerifySM3Password(password string, hashedPassword string) bool {
	idx := -1
	for i, c := range hashedPassword {
		if c == ':' {
			idx = i
			break
		}
	}
	if idx == -1 {
		return false
	}
	salt := hashedPassword[:idx]
	storedHash := hashedPassword[idx+1:]
	computedHash := SM3HashWithSalt(password, salt)
	return computedHash == storedHash
}
