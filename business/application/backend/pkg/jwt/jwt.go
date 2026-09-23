package jwt

import (
	"errors"
	"fmt"
	"time"

	"application/config"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// audience 用于区分两类 token。申报人 token 与管理端 token 共用同一个签名密钥，
// 若不校验 aud，申报人 token 会被 AdminAuth 当成管理端 token 接受（反之亦然）。
// 这两个常量是鉴权边界的一部分，切勿删除或改成同一个值。
const (
	AudienceUser  = "application-applicant"
	AudienceAdmin = "application-admin"
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
			Subject:   AudienceUser,
			Audience:  jwtlib.ClaimStrings{AudienceUser},
			ID:        uuid.New().String(),
		},
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

func ParseUserToken(tokenString string) (*UserClaims, error) {
	claims := &UserClaims{}
	if err := parseWithAudience(tokenString, claims, AudienceUser); err != nil {
		return nil, err
	}
	if claims.UserID == 0 {
		return nil, errors.New("invalid user token")
	}
	return claims, nil
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
			Subject:   AudienceAdmin,
			Audience:  jwtlib.ClaimStrings{AudienceAdmin},
			ID:        uuid.New().String(),
		},
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

func ParseAdminToken(tokenString string) (*AdminClaims, error) {
	claims := &AdminClaims{}
	if err := parseWithAudience(tokenString, claims, AudienceAdmin); err != nil {
		return nil, err
	}
	if claims.AdminID == 0 {
		return nil, errors.New("invalid admin token")
	}
	return claims, nil
}

// parseWithAudience 只接受 HS256、校验签发者与 aud。
// 用法白名单（WithValidMethods）必须保留：不加时 HS256/HS384/HS512 都会被接受，
// 算法混用会让攻击面无谓变大。
func parseWithAudience(tokenString string, claims jwtlib.Claims, audience string) error {
	cfg := config.Cfg.JWT
	token, err := jwtlib.ParseWithClaims(tokenString, claims,
		func(t *jwtlib.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwtlib.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(cfg.Secret), nil
		},
		jwtlib.WithValidMethods([]string{jwtlib.SigningMethodHS256.Alg()}),
		jwtlib.WithIssuer(cfg.Issuer),
		jwtlib.WithAudience(audience),
	)
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("invalid token")
	}
	return nil
}
