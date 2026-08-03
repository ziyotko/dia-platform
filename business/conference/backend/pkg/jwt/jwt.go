package jwt

import (
	"errors"
	"time"

	"conference/config"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// MemberClaims for member/frontend users
type MemberClaims struct {
	UserID      uint64 `json:"user_id"`
	Username    string `json:"username"`
	MemberLevel string `json:"member_level"`
	Branch      string `json:"branch"`
	jwtlib.RegisteredClaims
}

// AdminClaims for admin/backend users
type AdminClaims struct {
	AdminID  uint64 `json:"admin_id"`
	Username string `json:"username"`
	RoleCode string `json:"role_code"`
	jwtlib.RegisteredClaims
}

func GenerateMemberToken(userID uint64, username, memberLevel, branch string) (string, error) {
	cfg := config.Cfg.JWT
	claims := &MemberClaims{
		UserID:      userID,
		Username:    username,
		MemberLevel: memberLevel,
		Branch:      branch,
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

func ParseMemberToken(tokenString string) (*MemberClaims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, &MemberClaims{},
		func(t *jwtlib.Token) (interface{}, error) {
			return []byte(config.Cfg.JWT.Secret), nil
		})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*MemberClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid member token")
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
