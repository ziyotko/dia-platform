package jwt

import (
	"errors"
	"time"

	"base/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	TenantID uint64 `json:"tenant_id"`
	jti      string
	jwt.RegisteredClaims
}

func GenerateToken(userID uint64, username string, tenantID uint64) (string, error) {
	cfg := config.Cfg.JWT
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		TenantID: tenantID,
		jti:      uuid.NewString(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(cfg.ExpireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    cfg.Issuer,
			ID:        uuid.NewString(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
