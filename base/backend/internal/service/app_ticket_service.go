package service

import (
	"encoding/json"
	"errors"
	"time"

	"base/internal/models"
	"base/pkg/db"
	"base/pkg/redis"

	goredis "github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// AppTicketService 子应用一次性接入票据。
//
// 为什么需要：iframe 子应用过去通过 URL query 携带底座 access token（`base_token`），
// 会把长期凭证留在浏览器历史、Referer 与网关访问日志里。改成「一次性票据」后：
//   - 底座签发 60 秒有效、只用一次的随机票据，URL 里只带 `base_ticket`；
//   - 子应用启动时调 `POST .../auth/app-ticket/exchange`，用票据换回「access token + 用户信息」；
//   - 票据即使用日志/历史泄漏也已失效（取值与删除用 Lua 原子完成，不存在并发双花）。
const (
	appTicketPrefix = "auth:app-ticket:"
	// AppTicketTTL 票据有效期：够子应用加载后立刻兑换，又足够短
	AppTicketTTL = 60 * time.Second
)

var errAppTicketInvalid = errors.New("票据无效或已过期")

type AppTicketService struct{}

type appTicketPayload struct {
	UserID   uint64 `json:"userId"`
	IssuedAt int64  `json:"issuedAt"`
}

// Issue 为指定用户签发一次性票据，返回票据与有效期（秒）。
func (s AppTicketService) Issue(userID uint64, ttl time.Duration) (string, int, error) {
	if ttl <= 0 {
		ttl = AppTicketTTL
	}
	token, err := newOpaqueToken()
	if err != nil {
		return "", 0, err
	}
	payload, err := json.Marshal(appTicketPayload{UserID: userID, IssuedAt: time.Now().Unix()})
	if err != nil {
		return "", 0, err
	}
	if err := redis.Client.Set(redis.Ctx, appTicketPrefix+sha256Hex(token), payload, ttl).Err(); err != nil {
		return "", 0, err
	}
	return token, int(ttl.Seconds()), nil
}

// appTicketConsumeScript 原子「取值 + 删除」，保证同一票据只可能被消费一次。
var appTicketConsumeScript = goredis.NewScript(`
local v = redis.call('GET', KEYS[1])
if v then redis.call('DEL', KEYS[1]) end
return v
`)

// Consume 消费票据并返回对应用户（顺带校验账号仍为启用状态）。
func (s AppTicketService) Consume(ticket string) (*models.User, error) {
	if ticket == "" {
		return nil, errAppTicketInvalid
	}
	key := appTicketPrefix + sha256Hex(ticket)
	res, err := appTicketConsumeScript.Run(redis.Ctx, redis.Client, []string{key}).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, errAppTicketInvalid
		}
		logrus.WithError(err).Warn("消费应用票据失败")
		return nil, errRefreshStorage
	}
	raw, ok := res.(string)
	if !ok || raw == "" {
		return nil, errAppTicketInvalid
	}
	var payload appTicketPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, errAppTicketInvalid
	}
	var user models.User
	if err := db.DB.First(&user, payload.UserID).Error; err != nil {
		return nil, errAppTicketInvalid
	}
	if user.Status != 1 {
		return nil, errors.New("账号已禁用")
	}
	return &user, nil
}
