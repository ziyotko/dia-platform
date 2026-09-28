package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"server/config"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func GenerateToken(userID uint, email string, expiresHour int) (string, error) {
	if expiresHour <= 0 {
		expiresHour = config.AppConfig.JWT.ExpireHours
	}
	expireTime := time.Now().Add(time.Hour * time.Duration(expiresHour))
	// issuer 与 member / application 一样走配置（默认 caam-portal）；本服务解析时不校验 issuer
	issuer := config.AppConfig.JWT.Issuer
	if issuer == "" {
		issuer = "caam-portal"
	}
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
			Subject:   "user-token",
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWT.Secret))
}

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		// 只接受 HS256（与本服务签发算法一致），防止算法混淆攻击：
		// 原先只排除非 HMAC 算法，HS384/HS512 也会被放行。
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(config.AppConfig.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, err
}
