package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"portal/pkg/utils"
)

// memberLogoutBlacklistPrefix 外部会员登出黑名单的键前缀。
// 由 member 项目写入（member/internal/middleware/auth.go 的 logoutBlacklistPrefix、
// member/internal/service/auth_service.go 的 Logout），portal **只读**：
// `blacklist:<jti>` 存在即表示该 Token 已被会员主动登出。
// 改动该前缀必须同步 member 项目（这是本方案唯一的跨项目隐式契约）。
const memberLogoutBlacklistPrefix = "blacklist:"

// AuthExternalMemberMiddleware 校验「外部会员」令牌（business/member 项目签发的 JWT）。
//
// 场景：会员在会员中心（member）登录后，直接读取 portal 的「会员专区」已发布内容。
//
// 与 AuthMiddleware 的区别（**务必不要合并成一个中间件**）：
//   - 用 member 的密钥（jwt.member_secret）解析，强校验 issuer=member 的 issuer，只认 HS256；
//   - **不查 portal 的 user 表**：外部会员在 portal 没有账号，查表必然 401；
//   - 身份写入 externalMemberID / externalMemberName 与 utils.AuthSourceMember，
//     **不写 userID / currentUser**，控制器不得套用 portal 的用户/角色逻辑；
//   - 额外读 member 的登出黑名单（Redis `blacklist:<jti>`，即 1B-a 的吊销对齐），
//     使「会员登出后 portal 立即失效」。
//
// 该中间件只用于「会员专区」对外只读路由组：那里不挂防重放（member 前端不会带
// X-Request-Signature，挂了会 100% 拒绝）、不挂菜单校验（会员没有 portal 菜单授权）。
//
// 已知局限（选型时已确认）：member 的「改密 / 被禁用」只会写它自己库里的
// password_changed_at / token_invalid_before，portal 侧无法感知，
// 这类令牌最多可用到自然过期（member 默认 24h）。
func AuthExternalMemberMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			rejectExternalMember(c, "未提供会员登录凭证")
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			rejectExternalMember(c, "会员登录凭证格式错误")
			return
		}

		claims, err := utils.ParseMemberToken(parts[1])
		if err != nil {
			if errors.Is(err, utils.ErrExternalMemberDisabled) {
				// 未配置 jwt.member_secret：能力未启用，提示得具体一些，便于运维定位
				rejectExternalMember(c, "会员登录校验未启用，请联系管理员")
				return
			}
			rejectExternalMember(c, "会员登录已过期或无效，请重新登录会员中心")
			return
		}

		// 登出黑名单：Redis 不可用时 fail-closed（返回服务异常而不是放行），
		// 与 member 自身中间件的口径一致——缓存故障不能变成鉴权放开。
		if utils.RedisMember != nil && claims.ID != "" {
			_, err := utils.RedisMember.Get(utils.Ctx, memberLogoutBlacklistPrefix+claims.ID).Result()
			if err == nil {
				rejectExternalMember(c, "已退出登录，请重新登录会员中心")
				return
			}
			if !errors.Is(err, redis.Nil) {
				c.JSON(http.StatusOK, utils.Error(1, "服务暂时不可用，请稍后重试"))
				c.Abort()
				return
			}
		}

		utils.SetAuthSource(c, utils.AuthSourceMember)
		c.Set("externalMemberID", claims.MemberID)
		c.Set("externalMemberName", claims.Username)
		// 注意：claims.IsAdmin 是「member 后台管理员」标记，与 portal 管理员无关，
		// 因此这里不写入任何管理员标记，控制器也不得据此放行。
		c.Next()
	}
}

// rejectExternalMember 统一的鉴权失败出口：沿用 portal「HTTP 200 + 业务码 401」的约定。
func rejectExternalMember(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, utils.Error(AuthErrorCode, msg))
	c.Abort()
}
