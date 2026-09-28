package jwt

import (
	"member/config"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type MemberClaims struct {
	MemberID uint64 `json:"member_id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	jwtlib.RegisteredClaims
}

func GenerateToken(memberID uint64, username string, isAdmin bool) (string, error) {
	cfg := config.Cfg.JWT
	now := time.Now()
	claims := MemberClaims{
		MemberID: memberID,
		Username: username,
		IsAdmin:  isAdmin,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(now.Add(time.Duration(cfg.ExpireHours) * time.Hour)),
			IssuedAt:  jwtlib.NewNumericDate(now),
			Issuer:    cfg.Issuer,
			ID:        uuid.New().String(),
		},
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

func ParseToken(tokenString string) (*MemberClaims, error) {
	cfg := config.Cfg.JWT
	token, err := jwtlib.ParseWithClaims(tokenString, &MemberClaims{}, func(t *jwtlib.Token) (interface{}, error) {
		// 只接受 HS256：原先判 `*SigningMethodHMAC` 会同时放行 HS384/HS512（算法混淆面）。
		// 与 portal 的 utils/jwt.go 口径一致。
		if t.Method.Alg() != jwtlib.SigningMethodHS256.Alg() {
			return nil, jwtlib.ErrSignatureInvalid
		}
		return []byte(cfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*MemberClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwtlib.ErrSignatureInvalid
}
