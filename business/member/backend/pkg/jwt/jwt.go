package jwt

import (
	"member/config"
	"strings"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// DefaultIssuer member 未配置 jwt.issuer 时的兜底值。
// **必须与 portal 的 utils.MemberTokenIssuerDefault 保持一致**：portal 的 /member-zone/*
// 会强制校验 issuer，两侧不一致会让外部会员接口全部 401。
const DefaultIssuer = "business-member"

// defaultExpireHours 未配置 jwt.expire_hours 时的兜底有效期（小时）。
const defaultExpireHours = 24

// jwtIssuer 返回生效的 issuer。配置缺失时用兜底值，避免与 portal 的强校验不对称。
func jwtIssuer() string {
	if v := strings.TrimSpace(config.Cfg.JWT.Issuer); v != "" {
		return v
	}
	return DefaultIssuer
}

// jwtExpireHours 返回生效的有效期。配置为 0/负数时回退默认值，
// 否则会签发「立即过期」的 Token（登录接口返回成功，之后每个请求都 401）。
func jwtExpireHours() int {
	if h := config.Cfg.JWT.ExpireHours; h > 0 {
		return h
	}
	return defaultExpireHours
}

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
			ExpiresAt: jwtlib.NewNumericDate(now.Add(time.Duration(jwtExpireHours()) * time.Hour)),
			IssuedAt:  jwtlib.NewNumericDate(now),
			Issuer:    jwtIssuer(),
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
	},
		// 与 portal 的 ParseMemberToken 口径对齐：钉死算法、强制校验 issuer 与 exp。
		// 原先只验签名，凡共用密钥的签发方（如运维脚本）签出的 Token 都被接受，
		// 且无 exp 的 Token 在 member 侧永不过期。
		jwtlib.WithValidMethods([]string{jwtlib.SigningMethodHS256.Alg()}),
		jwtlib.WithIssuer(jwtIssuer()),
		jwtlib.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*MemberClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwtlib.ErrSignatureInvalid
}
