package jwt

import (
	"errors"
	"time"

	"application/config"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// UserClaims for applicant (frontend) users
type UserClaims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	jwtlib.RegisteredClaims
}

// AdminClaims for admin/backend users (manager, reviewer)
type AdminClaims struct {
	AdminID  uint64 `json:"admin_id"`
	Username string `json:"username"`
	RoleCode string `json:"role_code"`
	jwtlib.RegisteredClaims
}

func GenerateUserToken(userID uint64, username string) (string, error) {
	cfg := config.Cfg.JWT
	claims := &UserClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Duration(cfg.ExpireHours) * time.Hour)),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
			Issuer:    cfg.Issuer,
			ID:        uuid.New().String(),
		},
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

func ParseUserToken(tokenString string) (*UserClaims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, &UserClaims{},
		func(t *jwtlib.Token) (interface{}, error) {
			return []byte(config.Cfg.JWT.Secret), nil
		})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid user token")
}

func GenerateAdminToken(adminID uint64, username, roleCode string) (string, error) {
	cfg := config.Cfg.JWT
	claims := &AdminClaims{
		AdminID:  adminID,
		Username: username,
		RoleCode: roleCode,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Duration(cfg.ExpireHours) * time.Hour)),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
			Issuer:    cfg.Issuer,
			ID:        uuid.New().String(),
		},
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

func ParseAdminToken(tokenString string) (*AdminClaims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, &AdminClaims{},
		func(t *jwtlib.Token) (interface{}, error) {
			return []byte(config.Cfg.JWT.Secret), nil
		})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*AdminClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid admin token")
}
