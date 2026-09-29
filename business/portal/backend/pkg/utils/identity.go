package utils

import "github.com/gin-gonic/gin"

// 请求身份来源。portal 有两种完全不同的登录主体，控制器必须按来源分支判权：
//
//	portal —— 本系统自己的用户（middleware.AuthMiddleware 校验，写入 userID/currentUser）
//	member —— 外部会员（business/member 项目登录的会员，middleware.AuthExternalMemberMiddleware
//	          校验 member 的 JWT，写入 externalMemberID，**不写 userID/currentUser**）
//
// 两者刻意不共用任何上下文键：把会员身份当成 portal 用户会让操作日志、菜单校验、
// 归属判定（如 author_code）全部张冠李戴。
const (
	// AuthSourceKey gin.Context 中身份来源的键
	AuthSourceKey = "authSource"
	// AuthSourcePortal 本系统登录用户
	AuthSourcePortal = "portal"
	// AuthSourceMember 外部会员（member 项目）
	AuthSourceMember = "member"
)

// SetAuthSource 记录本次请求的身份来源（仅由鉴权中间件调用）。
func SetAuthSource(ctx *gin.Context, source string) {
	ctx.Set(AuthSourceKey, source)
}

// GetAuthSource 返回本次请求的身份来源；未经过鉴权中间件时为空串。
func GetAuthSource(ctx *gin.Context) string {
	v, ok := ctx.Get(AuthSourceKey)
	if !ok {
		return ""
	}
	source, _ := v.(string)
	return source
}

// IsExternalMember 本次请求是否为「外部会员」（member 项目的登录态）。
func IsExternalMember(ctx *gin.Context) bool {
	return GetAuthSource(ctx) == AuthSourceMember
}
