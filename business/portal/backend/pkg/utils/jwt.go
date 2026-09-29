package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"portal/config"
)

// ErrExternalMemberDisabled 未配置 member 密钥（jwt.member_secret / PORTAL_MEMBER_JWT_SECRET），
// 外部会员校验不可用。调用方应据此提示「未启用」，而不是笼统地报「令牌无效」。
var ErrExternalMemberDisabled = errors.New("外部会员登录校验未启用")

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

// ==================== 外部会员（business/member）令牌 ====================

// MemberTokenIssuerDefault member 未配置 issuer 时的兜底值（与 member config.yaml 的 jwt.issuer 一致）。
const MemberTokenIssuerDefault = "caam-member"

// ExternalMemberClaims 是 member 项目签发的 JWT 声明（member/pkg/jwt.MemberClaims）。
// portal 只借它确认「请求方是一个已登录的会员」，因此除 member_id 外一律不参与判权：
// member 的 is_admin 只是 member 后台的管理员标记，**绝不能**当作 portal 管理员使用。
type ExternalMemberClaims struct {
	MemberID uint64 `json:"member_id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// ParseMemberToken 校验 member 项目签发的登录令牌。
// 与 ParseToken 严格分开：使用 member 的密钥、强校验 issuer、强制存在 exp，
// 避免两套令牌（以及两套密钥）互相通用。
func ParseMemberToken(tokenString string) (*ExternalMemberClaims, error) {
	secret := config.AppConfig.JWT.MemberSecret
	if secret == "" {
		return nil, ErrExternalMemberDisabled
	}
	issuer := config.AppConfig.JWT.MemberIssuer
	if issuer == "" {
		issuer = MemberTokenIssuerDefault
	}
	token, err := jwt.ParseWithClaims(tokenString, &ExternalMemberClaims{}, func(token *jwt.Token) (any, error) {
		// 只接受 HS256（与 member 的签发算法一致），防止算法混淆攻击
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	}, jwt.WithIssuer(issuer))
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*ExternalMemberClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	// 必须带 exp：无过期时间的令牌一旦泄露即长期可用（member 的令牌总是带 exp）
	if claims.ExpiresAt == nil {
		return nil, jwt.ErrTokenRequiredClaimMissing
	}
	return claims, nil
}
